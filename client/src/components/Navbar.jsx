import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Navbar = () => {
    const navigate = useNavigate();
    const [userInfo, setUserInfo] = useState(null);
    const [users, setUsers] = useState([]);
    const [showModal, setShowModal] = useState(false);
    const [subject, setSubject] = useState('');
    const [body, setBody] = useState('');

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            fetchUserInfo(token);
        }
    }, []);

    const fetchUserInfo = async (token) => {
        try {
            const response = await fetch("http://localhost:8000/api/user-info", {
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
            const response = await fetch("http://localhost:8000/resend-activation-link", {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`
                }
            });
            handleLogout();
            // Ваша логика обработки ответа здесь
        } catch (error) {
            console.error("Error resending activation link:", error);
        }
    };

    const openModal = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch("http://localhost:8000/api/admin/users", {
                headers: {
                    Authorization: `Bearer ${token}`
                }
            });
            const data = await response.json();
            setUsers(data);
            setShowModal(true);
        } catch (error) {
            console.error("Error fetching users:", error);
        }
    };

    const handleSendMailing = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch("http://localhost:8000/api/admin/mailing", {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({ subject, body })
            });
            // Ваша логика обработки ответа здесь
        } catch (error) {
            console.error("Error sending mailing:", error);
        }
    };

    return (
        <div className="navbar" style={{ color: "white" }}>
            <div style={{ color: userInfo && !userInfo.isActivated ? "white" : "inherit" }}>
                Username: {userInfo && userInfo.username}
            </div>
            <div style={{ color: userInfo && !userInfo.isActivated ? "white" : "inherit" }}>
                {userInfo && userInfo.isActivated ? "" : <button onClick={handleResendActivation} style={{ color: "white" }}>Активируйте аккаунт</button>}
            </div>
            {userInfo && userInfo.ROLE === "ADMIN" ?
                <div>
                    <button onClick={openModal}>ADMIN ROLE</button>
                </div>
                : ""}
            <button onClick={handleLogout}>Logout</button>

            {showModal && (
                <div className="modal">
                    <div className="modal-content">
                        <div>
                            <label htmlFor="subject">Тема:</label>
                            <input type="text" id="subject" value={subject}
                                   onChange={(e) => setSubject(e.target.value)}/>
                        </div>
                        <div>
                            <label htmlFor="body">Текст сообщения:</label>
                            <textarea id="body" value={body} onChange={(e) => setBody(e.target.value)}/>
                        </div>
                        <button onClick={handleSendMailing}>Отправить рассылку</button>
                        <span className="close" onClick={() => setShowModal(false)}>&times;</span>
                        <h2>Почты пользователей:</h2>
                        <ul>
                            {users.map(user => (
                                <li key={user.ID}>{user.Email}</li>
                            ))}
                        </ul>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Navbar;
