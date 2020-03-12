const path = require('path');
const Dotenv = require('dotenv-webpack');
var HtmlWebpackPlugin = require('html-webpack-plugin');
const HtmlWebpackRootPlugin = require('html-webpack-root-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');

module.exports = {
  entry: './src/js/main.js',
  output: {
    filename: 'js/main.js',
    path: path.resolve(__dirname, 'built')
  },
  plugins: [
    new Dotenv({
      defaults: true,
    }),
    new HtmlWebpackPlugin({
       title:'Nebraska', 
       favicon:'built/favicon.png'
     }),
    new CleanWebpackPlugin(),
    new HtmlWebpackRootPlugin(),
    new CopyPlugin([
     { from:'public/fonts',to:'fonts' },
     { from:'public/favicon.png', to:''},
     { from:'public/img', to:'img'}
    ])
  ],
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      },
      {
        test: /\.(png|svg|jpg|gif)$/,
        use: ['file-loader']
      },
      {
        test: /\.(woff|woff2|eot|ttf|otf)$/,
        use:['file-loader']
      }
    ]
  },
  optimization: {
    minimize: true
  },
  mode: 'production'
};
