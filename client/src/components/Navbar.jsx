import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Navbar = () => {
    const navigate = useNavigate();
    const [userInfo, setUserInfo] = useState(null);

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            fetchUserInfo(token);
        }
    }, []);

    const fetchUserInfo = async (token) => {
        try {
            const response = await fetch("http://103.13.211.78:8000/api/user-info", {
                headers: {
                    Authorization: `Bearer ${token}`
                }
            });
            const data = await response.json();
            setUserInfo(data);
        } catch (error) {
            console.error("Error fetching user info:", error);
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('token');
        window.history.go("/login")
    };

    const handleResendActivation = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch("http://103.13.211.78:8000/resend-activation-link", {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`
                }
            });
            handleLogout()
            // Ваша логика обработки ответа здесь
        } catch (error) {
            console.error("Error resending activation link:", error);
        }
    };

    return (
        <div className="navbar" style={{color: "white"}}>
            <div style={{ color: userInfo && !userInfo.isActivated ? "white" : "inherit" }}>
                Username: {userInfo && userInfo.username}
            </div>
            <div style={{ color: userInfo && !userInfo.isActivated ? "white" : "inherit" }}>
                {userInfo && userInfo.isActivated ? "" : <button onClick={handleResendActivation} style={{ color: "white" }}>Активируйте аккаунт</button>}
            </div>
            <button onClick={handleLogout}>Logout</button>
        </div>
    );
};

export default Navbar;
