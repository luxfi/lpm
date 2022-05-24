package apm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/leveldb"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/perms"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/avalanche-plugins-core/core"

	"github.com/ava-labs/apm/admin"
	"github.com/ava-labs/apm/engine"
	"github.com/ava-labs/apm/git"
	"github.com/ava-labs/apm/repository"
	"github.com/ava-labs/apm/types"
	"github.com/ava-labs/apm/url"
	"github.com/ava-labs/apm/util"
)

var (
	dbDir         = "db"
	repositoryDir = "repositories"
	tmpDir        = "tmp"

	repoPrefix         = []byte("repo")
	vmPrefix           = []byte("vm")
	installedVMsPrefix = []byte("installed_vms")
	globalPrefix       = []byte("global")
)

type Config struct {
	Directory        string
	Auth             http.BasicAuth
	AdminApiEndpoint string
	PluginDir        string
}

type APM struct {
	repositoriesPath string
	tmpPath          string
	pluginPath       string

	db           database.Database // base db
	repositoryDB database.Database // repositories we track
	installedVMs database.Database // vms that are currently installed

	globalRegistry repository.Registry // all vms and subnets able to be installed

	auth http.BasicAuth

	adminClient admin.Client
	httpClient  url.Client

	engine engine.Engine
}

func New(config Config) (*APM, error) {
	dbDir := filepath.Join(config.Directory, dbDir)
	db, err := leveldb.New(dbDir, []byte{}, logging.NoLog{})
	if err != nil {
		return nil, err
	}

	a := &APM{
		repositoriesPath: filepath.Join(config.Directory, repositoryDir),
		tmpPath:          filepath.Join(config.Directory, tmpDir),
		pluginPath:       config.PluginDir,
		db:               db,
		globalRegistry: repository.NewRegistry(repository.RegistryConfig{
			Alias: globalPrefix,
			DB:    db,
		}),
		repositoryDB: prefixdb.New(repoPrefix, db),
		installedVMs: prefixdb.New(installedVMsPrefix, db),
		auth:         config.Auth,
		adminClient: admin.NewHttpClient(
			admin.HttpClientConfig{
				Endpoint: fmt.Sprintf("http://%s", config.AdminApiEndpoint),
			},
		),
		httpClient: url.NewHttpClient(),
		engine:     engine.NewWorkflowEngine(),
	}

	if err := os.MkdirAll(a.repositoriesPath, perms.ReadWriteExecute); err != nil {
		return nil, err
	}

	//TODO simplify this
	coreKey := []byte(core.Alias)
	if _, err = a.repositoryDB.Get(coreKey); err == database.ErrNotFound {
		err := a.AddRepository(core.Alias, core.URL)
		if err != nil {
			return nil, err
		}
	}

	repoMetadata, err := a.repositoryMetadataFor(coreKey)
	if err != nil {
		return nil, err
	}

	if repoMetadata.Commit == plumbing.ZeroHash {
		fmt.Println("Bootstrap not detected. Bootstrapping...")
		err := a.Update()
		if err != nil {
			return nil, err
		}

		fmt.Println("Finished bootstrapping.")
	}
	return a, nil
}

func parseAndRun(alias string, globalRegistry database.Database, command func(string) error) error {
	if qualifiedName(alias) {
		return command(alias)
	}

	fullName, err := getFullNameForAlias(globalRegistry, alias)
	if err != nil {
		return err
	}

	return command(fullName)

}

func (a *APM) Install(alias string) error {
	return parseAndRun(alias, a.globalRegistry.VMs(), a.install)
}

func (a *APM) install(name string) error {
	nameBytes := []byte(name)

	ok, err := a.installedVMs.Has(nameBytes)
	if err != nil {
		return err
	}

	if ok {
		fmt.Printf("VM %s is already installed. Skipping.\n", name)
		return nil
	}

	repoAlias, plugin := util.ParseQualifiedName(name)
	organization, repo := util.ParseAlias(repoAlias)

	workflow := engine.NewInstallWorkflow(engine.InstallWorkflowConfig{
		Name:         name,
		Plugin:       plugin,
		Organization: organization,
		Repo:         repo,
		TmpPath:      a.tmpPath,
		PluginPath:   a.pluginPath,
		InstalledVMs: a.installedVMs,
		Registry: repository.NewRegistry(repository.RegistryConfig{
			Alias: []byte(repoAlias),
			DB:    a.db,
		}),
		HttpClient: a.httpClient,
	})

	return a.engine.Execute(workflow)
}

func (a *APM) Uninstall(alias string) error {
	return parseAndRun(alias, a.globalRegistry.VMs(), a.uninstall)
}

func (a *APM) uninstall(name string) error {
	nameBytes := []byte(name)

	ok, err := a.installedVMs.Has(nameBytes)
	if err != nil {
		return err
	}

	if !ok {
		fmt.Printf("VM %s is already not installed. Skipping.\n", name)
		return nil
	}

	alias, plugin := util.ParseQualifiedName(name)

	repoDB := prefixdb.New([]byte(alias), a.db)
	repoVMDB := prefixdb.New(vmPrefix, repoDB)

	ok, err = repoVMDB.Has([]byte(plugin))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("Virtual machine already %s doesn't exist in the vm registry for %s.", name, alias)
		return nil
	}

	if err := a.installedVMs.Delete([]byte(plugin)); err != nil {
		return err
	}

	fmt.Printf("Successfully uninstalled %s.", name)

	return nil
}

func (a *APM) JoinSubnet(alias string) error {
	return parseAndRun(alias, a.globalRegistry.Subnets(), a.joinSubnet)
}

func (a *APM) joinSubnet(fullName string) error {
	alias, plugin := util.ParseQualifiedName(fullName)
	aliasBytes := []byte(alias)
	repoRegistry := repository.NewRegistry(repository.RegistryConfig{
		Alias: aliasBytes,
		DB:    a.db,
	})

	subnetBytes, err := repoRegistry.Subnets().Get([]byte(plugin))
	if err != nil {
		return err
	}

	definition := &repository.Definition[types.Subnet]{}
	if err := yaml.Unmarshal(subnetBytes, definition); err != nil {
		return err
	}
	subnet := definition.Definition

	// TODO prompt user, add force flag
	fmt.Printf("Installing virtual machines for subnet %s.\n", subnet.ID())
	for _, vm := range subnet.VMs_ {
		if err := a.Install(vm); err != nil {
			return err
		}
	}

	fmt.Printf("Updating virtual machines...\n")
	if err := a.adminClient.LoadVMs(); err != nil {
		return err
	}

	fmt.Printf("Whitelisting subnet %s...\n", subnet.ID())
	if err := a.adminClient.WhitelistSubnet(subnet.ID()); err != nil {
		return err
	}

	fmt.Printf("Finished installing virtual machines for subnet %s.\n", subnet.ID_)
	return nil
}

func (a *APM) Upgrade(alias string) error {
	return nil
}

func (a *APM) Search(alias string) error {
	return nil
}

func (a *APM) Info(alias string) error {
	if qualifiedName(alias) {
		return a.install(alias)
	}

	fullName, err := getFullNameForAlias(a.globalRegistry.VMs(), alias)
	if err != nil {
		return err
	}

	return a.info(fullName)
}

func (a *APM) info(fullName string) error {
	return nil
}

func (a *APM) Update() error {
	itr := a.repositoryDB.NewIterator()

	for itr.Next() {
		aliasBytes := itr.Key()
		organization, repo := util.ParseAlias(string(aliasBytes))

		repositoryMetadata, err := a.repositoryMetadataFor(aliasBytes)
		if err != nil {
			return err
		}
		repositoryPath := filepath.Join(a.repositoriesPath, organization, repo)
		gitRepo, err := git.NewRemote(repositoryMetadata.URL, repositoryPath, "refs/heads/main", &a.auth)
		if err != nil {
			return err
		}

		previousCommit := repositoryMetadata.Commit
		latestCommit, err := gitRepo.Head()
		if err != nil {
			return err
		}

		if latestCommit == previousCommit {
			fmt.Printf("Already at latest for %s@%s.\n", repo, latestCommit)
			continue
		}

		workflow := engine.NewUpdateWorkflow(engine.UpdateWorkflowConfig{
			Engine:         a.engine,
			RepoName:       repo,
			RepositoryPath: repositoryPath,
			AliasBytes:     aliasBytes,
			PreviousCommit: previousCommit,
			LatestCommit:   latestCommit,
			RepoRegistry: repository.NewRegistry(repository.RegistryConfig{
				Alias: aliasBytes,
				DB:    a.db,
			}),
			GlobalRegistry:     a.globalRegistry,
			RepositoryMetadata: *repositoryMetadata,
			RepositoryDB:       a.repositoryDB,
			InstalledVMs:       a.installedVMs,
			DB:                 a.db,
			TmpPath:            a.tmpPath,
			PluginPath:         a.pluginPath,
			HttpClient:         a.httpClient,
		})

		if err := a.engine.Execute(workflow); err != nil {
			return err
		}
	}

	return nil
}

func (a *APM) repositoryMetadataFor(alias []byte) (*repository.Metadata, error) {
	repositoryMetadataBytes, err := a.repositoryDB.Get(alias)
	if err != nil && err != database.ErrNotFound {
		return nil, err
	}

	repositoryMetadata := &repository.Metadata{}
	if err := yaml.Unmarshal(repositoryMetadataBytes, repositoryMetadata); err != nil {
		return nil, err
	}

	return repositoryMetadata, nil
}

func (a *APM) AddRepository(alias string, url string) error {
	aliasBytes := []byte(alias)
	ok, err := a.repositoryDB.Has(aliasBytes)
	if err != nil {
		return err
	}
	if ok {
		fmt.Printf("%s is already registered as a repository.\n", alias)
		return nil
	}

	metadata := repository.Metadata{
		Alias:  alias,
		URL:    url,
		Commit: plumbing.ZeroHash, // hasn't been synced yet
	}
	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	return a.repositoryDB.Put(aliasBytes, metadataBytes)
}

func (a *APM) RemoveRepository(alias string) error {
	if qualifiedName(alias) {
		return a.removeRepository(alias)
	}

	fullName, err := getFullNameForAlias(a.globalRegistry.VMs(), alias)
	if err != nil {
		return err
	}

	return a.removeRepository(fullName)
}

func (a *APM) removeRepository(name string) error {
	if name == core.Alias {
		fmt.Printf("Can't remove %s (required repository).\n", core.Alias)
		return nil
	}

	//TODO don't let people remove core
	aliasBytes := []byte(name)

	repoRegistry := repository.NewRegistry(repository.RegistryConfig{
		Alias: aliasBytes,
		DB:    a.db,
	})

	// delete all the plugin definitions in the repository
	vmItr := repoRegistry.VMs().NewIterator()
	for vmItr.Next() {
		if err := repoRegistry.VMs().Delete(vmItr.Key()); err != nil {
			return err
		}
	}

	subnetItr := repoRegistry.VMs().NewIterator()
	for subnetItr.Next() {
		if err := repoRegistry.VMs().Delete(subnetItr.Key()); err != nil {
			return err
		}
	}

	// remove it from our list of tracked repositories
	return a.repositoryDB.Delete(aliasBytes)
}

func (a *APM) ListRepositories() error {
	itr := a.repositoryDB.NewIterator()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "alias\turl")
	for itr.Next() {
		metadata := &repository.Metadata{}
		if err := yaml.Unmarshal(itr.Value(), metadata); err != nil {
			return err
		}

		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", metadata.Alias, metadata.URL))
	}
	w.Flush()
	return nil
}

func qualifiedName(name string) bool {
	parsed := strings.Split(name, ":")
	return len(parsed) > 1
}

func getFullNameForAlias(db database.Database, alias string) (string, error) {
	bytes, err := db.Get([]byte(alias))
	if err != nil {
		return "", err
	}

	registry := &repository.List{}
	if err := yaml.Unmarshal(bytes, registry); err != nil {
		return "", err
	}

	if len(registry.Repositories) > 1 {
		return "", errors.New(fmt.Sprintf("more than one match found for %s. Please specify the fully qualified name. Matches: %s.\n", alias, registry.Repositories))
	}

	return fmt.Sprintf("%s:%s", registry.Repositories[0], alias), nil
}
