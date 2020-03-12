const path = require('path');
const Dotenv = require('dotenv-webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
module.exports = {
  entry: './src/js/main.js',
  output: {
    filename: 'main.js',
    path: path.resolve(__dirname, 'public')
  },
  plugins: [
    new Dotenv({
      defaults: true,
    }),
    new HtmlWebpackPlugin({
      template: "public/index.html"
    })
  ],
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      }
    ]
  },
  devtool: 'eval-cheap-module-source-map',
  devServer:{
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8000'
    },
    contentBase:path.join(__dirname, "/public"),
    hot: true,
    watchOptions: {
        poll: true
    }
  }
};
