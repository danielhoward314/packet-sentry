import { useAxiosInterceptor } from "@/hooks/useAxiosInterceptor";

export const AxiosInterceptorProvider = () => {
  useAxiosInterceptor();
  return null;
};
