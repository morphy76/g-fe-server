// Generated using webpack-cli https://github.com/webpack/webpack-cli

import webpack from "webpack";
import path from "path";
import { fileURLToPath } from "url";
import { CleanWebpackPlugin } from "clean-webpack-plugin";
import ESLintPlugin from "eslint-webpack-plugin";
import HtmlWebpackPlugin from "html-webpack-plugin";
import MiniCssExtractPlugin from "mini-css-extract-plugin";
import WorkboxWebpackPlugin from "workbox-webpack-plugin";
import { Configuration as DevServerConfiguration } from "webpack-dev-server";
import { reactCompilerLoader } from "react-compiler-webpack";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const __production = process.env["NODE_ENV"] === "production";

const stylesHandler = __production
  ? MiniCssExtractPlugin.loader
  : "style-loader";

const config: webpack.Configuration & { devServer: DevServerConfiguration } = {
  entry: "./src/index.tsx",
  output: {
    path: path.resolve(__dirname, "dist"),
    // publicPath: "ui",
    filename: "[name]-[fullhash].js",
    clean: true,
  },
  devServer: {
    open: true,
    host: "localhost",
  },
  watchOptions: {
    ignored: /node_modules/,
  },
  module: {
    rules: [
      {
        test: /\.(ts|tsx)$/i,
        use: [
          {
            loader: "ts-loader",
            options: {
              compilerOptions: {
                declaration: !__production,
                declarationMap: !__production
              }
            }
          },
          {
            loader: reactCompilerLoader,
          },
        ],
        exclude: ["/node_modules/"],
      },
      {
        test: /\.css$/i,
        use: [stylesHandler, "css-loader"],
      },
      {
        test: /\.s[ac]ss$/i,
        use: [stylesHandler, "css-loader", "sass-loader"],
      },
      {
        test: /\.(eot|svg|ttf|woff|woff2|png|jpg|gif)$/i,
        type: "asset",
      },

      // Add your rules for custom modules here
      // Learn more about loaders from https://webpack.js.org/loaders/
    ],
  },
  resolve: {
    extensions: [".tsx", ".ts", ".jsx", ".js"],
  },
};

config.plugins = [
  new CleanWebpackPlugin(),
  new HtmlWebpackPlugin({
    template: "../public/index.html",
    filename: "index.html",
    minify: __production,
    title: "Q-FE" + (__production ? " - Production" : " - Development"),
    favicon: "../public/favicon.ico",
  }),
  new ESLintPlugin({
    overrideConfigFile: "./.eslintrc.json",
  }),
];

if (__production) {
  config.mode = "production";
  config.plugins.push(new MiniCssExtractPlugin({
    filename: "[name]-[fullhash].css",
    chunkFilename: "[id]-[fullhash].css",
  }));
  config.plugins.push(
    new WorkboxWebpackPlugin.GenerateSW({
      clientsClaim: true,
      skipWaiting: true,
    })
  );
  config.optimization = {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          chunks: 'all',
        },
      },
    },
  };
} else {
  config.mode = "development";
  config.devtool = "source-map";
}

export default config;
