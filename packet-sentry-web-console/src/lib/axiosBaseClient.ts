import axios, { AxiosInstance } from "axios";

const baseClient: AxiosInstance = axios.create({
  baseURL: "",
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: false,
});

export default baseClient;
