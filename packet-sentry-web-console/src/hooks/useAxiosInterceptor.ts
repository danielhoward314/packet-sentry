import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useEnv } from "@/contexts/EnvContext";
import baseClient from "@/lib/axiosBaseClient";
import { LOCALSTORAGE } from "@/lib/consts";
import {
    AxiosRequestConfig,
    AxiosError,
    AxiosResponse,
    InternalAxiosRequestConfig
} from "axios";

interface RetryAxiosRequestConfig extends AxiosRequestConfig {
    _retry?: boolean;
  }

export function useAxiosInterceptor() {
  const { API_BASE_URL } = useEnv();
  const { API_ACCESS_TOKEN, API_REFRESH_TOKEN } = LOCALSTORAGE;
  const navigate = useNavigate();

  useEffect(() => {
    baseClient.defaults.baseURL = `${API_BASE_URL}/v1`;

    const requestInterceptor = baseClient.interceptors.request.use(
        async (config: InternalAxiosRequestConfig & { _retry?: boolean }) => {
        const token = localStorage.getItem(API_ACCESS_TOKEN);
        if (token) {
          config.headers = config.headers || {};
          config.headers["Authorization"] = `Bearer ${token}`;
        }
        return config;
      },
      (error: AxiosError) => Promise.reject(error)
    );

    const responseInterceptor = baseClient.interceptors.response.use(
      (response: AxiosResponse) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as RetryAxiosRequestConfig;
        if (
            error.response?.status === 401 &&
            !originalRequest._retry &&
            !originalRequest.url?.includes("/refresh")
        ) {
          originalRequest._retry = true;

          try {
            const refreshToken = localStorage.getItem(API_REFRESH_TOKEN);
            const refreshFormData = {
              jwt: refreshToken,
              claimsType: 2, // ClaimsType_API_AUTHORIZATION
            };

            // Refresh token request
            const refreshResponse = await baseClient.post(
              '/refresh',
              refreshFormData
            );

            const newToken = refreshResponse.data?.jwt;
            if (!newToken) {
              navigate("/login");
              return;
            }

            localStorage.setItem(API_ACCESS_TOKEN, newToken);
            originalRequest.headers = {
              ...originalRequest.headers,
              Authorization: `Bearer ${newToken}`,
            };

            // Retry the original request with new token
            return baseClient(originalRequest);
          } catch (err) {
            navigate("/login");
            return Promise.reject(err);
          }
        }

        return Promise.reject(error);
      }
    );

    // Cleanup interceptors when component unmounts
    return () => {
      baseClient.interceptors.request.eject(requestInterceptor);
      baseClient.interceptors.response.eject(responseInterceptor);
    };
  }, [navigate]);
}
