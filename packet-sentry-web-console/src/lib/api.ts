import {
  ActivateAdministratorRequest,
  CreateAdministratorRequest,
  UpdateAdministratorRequest,
  UpdateOrganizationRequest,
} from "@/types/api";
import baseClient from "./axiosBaseClient";
import { AxiosResponse } from "axios";

export async function getOrganization(id: string): Promise<any> {
  const res = await baseClient.get(`/organizations/${id}`);
  return res.data;
}

export async function updateOrganization(
  id: string,
  request: UpdateOrganizationRequest,
): Promise<AxiosResponse<void>> {
  return baseClient.put(`/organizations/${id}`, request);
}

export async function getAdministrator(id: string): Promise<any> {
  const res = await baseClient.get(`/administrators/${id}`);
  return res.data;
}

export async function updateAdministrator(
  id: string,
  request: UpdateAdministratorRequest,
): Promise<AxiosResponse<void>> {
  return baseClient.put(`/administrators/${id}`, request);
}

export async function listAdministrators(organizationId: string): Promise<any> {
  const res = await baseClient.get(
    `/administrators?organizationId=${organizationId}`,
  );
  return res.data;
}

export async function createAdministrator(
  request: CreateAdministratorRequest,
): Promise<AxiosResponse<void>> {
  return baseClient.post(`/administrators`, request);
}

export async function activateAdministrator(
  request: ActivateAdministratorRequest,
): Promise<AxiosResponse<void>> {
  return baseClient.post("/activate", request);
}
