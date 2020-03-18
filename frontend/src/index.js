import React from 'react';
import ReactDOM from 'react-dom';
import AppRoutes from './js/App';
import { BrowserRouter } from 'react-router-dom';

ReactDOM.render(<BrowserRouter><AppRoutes /></BrowserRouter>, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA

