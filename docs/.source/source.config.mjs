// source.config.ts
import {
  defineConfig,
  defineDocs
} from "fumadocs-mdx/config";
import rehypePrettyCode from "rehype-pretty-code";
var source_config_default = defineConfig({
  mdxOptions: {
    rehypePlugins: [
      [
        rehypePrettyCode,
        {
          theme: {
            dark: "github-dark-dimmed",
            light: "github-light"
          },
          keepBackground: false,
          defaultLang: "go"
        }
      ]
    ]
  }
});
var docs = defineDocs({
  dir: "content/docs"
});
export {
  source_config_default as default,
  docs
};
