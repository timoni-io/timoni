import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { NaiveUiResolver } from "unplugin-vue-components/resolvers";
import path from "path";
import { vueI18n } from "@intlify/vite-plugin-vue-i18n";
// import { createI18n } from "vue-i18n";
import * as icons from "@mdi/js";
import monacoEditorPlugin from "vite-plugin-monaco-editor";
import pluginRewriteAll from "vite-plugin-rewrite-all";

// https://vitejs.dev/config/
export default ({ mode }) => {
  const env = { ...process.env, ...loadEnv(mode, process.cwd()) };
  return defineConfig({
    server: {
      proxy: {
        "/api": {
          target: env.VITE_API_PROTOCOL + "://" + env.VITE_API_HOST,
          changeOrigin: true,
          secure: false,
        },
        "/cli": {
          target: env.VITE_API_PROTOCOL + "://" + env.VITE_API_HOST,
          changeOrigin: true,
          secure: false,
        },
        "/term": {
          target: env.VITE_API_PROTOCOL + "://" + env.VITE_API_HOST,
          changeOrigin: true,
          secure: false,
        },
        "/logo.svg": {
          target: env.VITE_API_PROTOCOL + "://" + env.VITE_API_HOST,
          changeOrigin: true,
          secure: false,
        },
      },
    },
    plugins: [
      vue({
        reactivityTransform: true,
      }),
      AutoImport({
        imports: [
          "vue",
          "@vueuse/core",
          "vue-i18n",
          {
            zod: ["z"],
            "@/zodios/api": ["api", "ResType"],
            "@mdi/js": Object.keys(icons).filter((i) => i !== "default"),
            "@/utils/is": ["is"],
          },
        ],
        vueTemplate: true,
      }),
      Components({
        resolvers: [NaiveUiResolver()],
        dts: true,
      }),
      vueI18n({
        include: path.resolve(__dirname, "./src/locales.json"),
      }),
      monacoEditorPlugin,
      pluginRewriteAll(),
    ],
    resolve: {
      alias: {
        "@/": `${path.resolve(__dirname, "src")}/`,
        "vue-i18n": "vue-i18n/dist/vue-i18n.runtime.esm-bundler.js",
      },
    },
    build: {
      target: "esnext",
      // chunkSizeWarningLimit: 5000,
      // rollupOptions: {
      //   output: {
      //     manualChunks(id) {
      //       if (id.includes('node_modules')) {
      //         return id.toString().split('node_modules/')[1].split('/')[0].toString();
      //       }
      //     }
      //   }
      // }
    },
    optimizeDeps: {
      esbuildOptions: {
        target: "esnext",
      },
    },
    esbuild: {
      target: "esnext",
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `@import "./src/assets/scss/main.scss";`,
        },
      },
    },
  });
};
