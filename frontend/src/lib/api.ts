import axios from 'axios';

// Create an Axios instance
const api = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true, // Important for cookies (refresh_token)
});

// Request interceptor to add Bearer token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response interceptor for handling token refresh
let isRefreshing = false;
let failedQueue: any[] = [];

const processQueue = (error: any, token: string | null = null) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(token);
        }
    });

    failedQueue = [];
};

api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;

        // If error is 401 and we haven't retried yet, AND it's not the refresh endpoint itself
        if (error.response?.status === 401 && !originalRequest._retry && !originalRequest.url?.includes('/auth/refresh')) {

            if (isRefreshing) {
                return new Promise(function (resolve, reject) {
                    failedQueue.push({ resolve, reject });
                }).then(token => {
                    originalRequest.headers['Authorization'] = 'Bearer ' + token;
                    return api(originalRequest);
                }).catch(err => {
                    return Promise.reject(err);
                });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                // Attempt to refresh token
                const { data } = await api.post('/auth/refresh');

                // Save new access token
                localStorage.setItem('token', data.token);

                // Update header for retry
                originalRequest.headers.Authorization = `Bearer ${data.token}`;

                processQueue(null, data.token);
                isRefreshing = false;

                // Retry original request
                return api(originalRequest);
            } catch (refreshError) {
                processQueue(refreshError, null);
                isRefreshing = false;

                // If refresh fails, clear token and redirect
                localStorage.removeItem('token');
                if (typeof window !== 'undefined') {
                    window.location.href = '/login';
                }
                // Return valid promise that never resolves to stop downstream error handling from triggering retries
                return new Promise(() => { });
            }
        }

        if (error.response?.status === 403) {
            // Account suspended or deleted
            localStorage.removeItem('token');
            if (typeof window !== 'undefined') {
                const errorMessage = error.response?.data?.error || 'Access denied';
                window.location.href = `/login?error=${encodeURIComponent(errorMessage)}`;
            }
            return new Promise(() => { });
        }

        return Promise.reject(error);
    }
);

export default api;
