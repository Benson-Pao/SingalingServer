class TokenManager {
    static getToken() {
        return localStorage.getItem('token'); // 獲取當前存儲的 token
    }

    static getRefreshToken() {
        return localStorage.getItem('refreshToken'); // 獲取 refresh token
    }

    static setToken(token) {
        localStorage.setItem('token', token); // 存儲新的 token
    }

    static setRefreshToken(refreshToken) {
        localStorage.setItem('refreshToken', refreshToken); // 存儲新的 refresh token
    }

    // 檢查 token 是否過期
    static isTokenExpired(token) {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const expiration = payload.exp;
        return Date.now() >= expiration * 1000; // 檢查是否過期
    }

    // 使用 refresh token 刷新新的 access token
    static async refreshToken(refreshToken) {
        const response = await fetch('/api/refresh-token', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refreshToken })
        });
        const data = await response.json();
        if (response.ok) {
            this.setToken(data.token);
            this.setRefreshToken(data.refreshToken);
            return data.token;
        } else {
            throw new Error('Token refresh failed');
        }
    }
}

export default TokenManager;
