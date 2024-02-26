import React, {useEffect, useState} from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate } from "react-router-dom";
import './App.css';
import Home from './pages/Home';
import Login from './components/Login';
import Signup from './components/Signup';

const App = () => {
    const token = localStorage.getItem('token');

    return (
        <Router>
            <Routes>
                <Route path="/" element={token ? <Home /> : <Navigate to="/login" />} />
                <Route path="/login" element={!token ? <Login /> : <Navigate to="/" />} />
                <Route path="/signup" element={!token ? <Signup /> : <Navigate to="/" />} />
            </Routes>
        </Router>
    );
};

export default App;
