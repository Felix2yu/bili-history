/** @type {import("prettier").Config} */
const config = {
  useTabs: false,
  tabWidth: 2,
  singleQuote: true,
  trailingComma: 'none',
  printWidth: 100,
  plugins: ['prettier-plugin-tailwindcss'],
  tailwindStylesheet: './assets/css/main.css'
};

export default config;
