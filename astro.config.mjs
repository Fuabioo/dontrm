// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import catppuccin from "@catppuccin/starlight";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "dontrm",
      logo: {
        src: "./src/assets/logo.svg"
      },
      pagefind: false,
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/Fuabioo/dontrm"
        }
      ],
      plugins: [
        catppuccin({
          dark: { flavor: "mocha" },
          light: { flavor: "latte" }
        })
      ]
    })
  ]
});
