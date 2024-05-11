// Generated using webpack-cli https://github.com/webpack/webpack-cli

const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const ESLintPlugin = require("eslint-webpack-plugin");
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const json5 = require('json5');

const isProduction = process.env.NODE_ENV == 'production';

const config = {
  entry: './src/index.tsx',
  output: {
    path: path.resolve(__dirname, '../../../target/frontend/'),
  },
  watchOptions: {
    ignored: /node_modules/,
  },
  plugins: [
    new CleanWebpackPlugin(),
    new HtmlWebpackPlugin({
      template: '../public/index.html',
      filename: 'index.html',
      minify: isProduction,
      title: 'Q-FE' + (isProduction ? ' - Production' : ' - Development'),
      favicon: '../public/favicon.ico',
    }),
    new ForkTsCheckerWebpackPlugin({
      async: false
    }),
    new ESLintPlugin({
      extensions: ["js", "jsx", "ts", "tsx"],
    }),
  ],
  module: {
    rules: [
      {
        test: /\.(ts|tsx)$/i,
        loader: 'ts-loader',
        exclude: ['/node_modules/'],
      },
      {
        test: /\.s[ac]ss$/i,
        use: [
          isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
          {
            loader: 'css-loader',
            options: {
              modules: {
                mode: 'local',
                localIdentName: isProduction ? '[hash:base64]' : '[local]',
              },
            },
          },
          "sass-loader",
        ],
      },
      {
        test: /\.(eot|svg|ttf|woff|woff2|png|jpg|gif)$/i,
        type: 'asset',
      },
      {
        test: /\.bundle.json$/i,
        type: 'json',
        parser: {
          parse: json5.parse,
        },
      },
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.jsx', '.js'],
    plugins: [new TsconfigPathsPlugin()],
  },
  optimization: {
    splitChunks: {
      cacheGroups: {
        axiosVendor: {
          test: /[\\/]node_modules[\\/](axios)[\\/]/,
          chunks: 'all',
        },
        reactVendor: {
          test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
          chunks: 'all',
        },
        reactIntlVendor: {
          test: /[\\/]node_modules[\\/](react-intl)[\\/]/,
          chunks: 'all',
        },
        reactQuerylVendor: {
          test: /[\\/]node_modules[\\/](react-query)[\\/]/,
          chunks: 'all',
        },
      },
      chunks: 'all',
    },
  },
};

module.exports = () => {
  if (isProduction) {
    config.mode = 'production';
    config.plugins.push(new MiniCssExtractPlugin());
    console.log('Production mode');
  } else {
    config.mode = 'development';
    config.devtool = 'source-map';
    console.log('Development mode');
  }
  return config;
};
