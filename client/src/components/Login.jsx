import React from "react";
import { Link } from "react-router-dom";
import "./Auth.css";

const Login = () => {
  return (
    <div className="auth-container">
      <h1>Log in</h1>
      <form>
        <div className="form-group">
          <label htmlFor="email">Email:</label>
          <input type="email" id="email" name="email" required />
        </div>
        <div className="form-group last">
          <label htmlFor="password">Password:</label>
          <input type="password" id="password" name="password" required />
        </div>
      </form>
      <button className="btn-auth" type="submit">
        Log in
      </button>
      <p>
        Don't have an account? <a href="/signup">Sign up</a>
      </p>
    </div>
  );
};

export default Login;
