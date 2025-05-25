import {
  ActivateAdministratorRequest,
  CreateAdministratorRequest,
  CreateInstallKeyRequest,
  UpdateAdministratorRequest,
  UpdateDeviceRequest,
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

export async function createInstallKey(
  request: CreateInstallKeyRequest,
): Promise<any> {
  const response = await baseClient.post("/install-keys", request);
  return response.data;
}

export async function listDevices(organizationId: string): Promise<any> {
  const res = await baseClient.get(`/devices?organizationId=${organizationId}`);
  return res.data;
}

export async function getDevice(id: string): Promise<any> {
  const res = await baseClient.get(`/devices/${id}`);
  return res.data;
}

export async function updateDevice(
  id: string,
  request: UpdateDeviceRequest,
): Promise<AxiosResponse<void>> {
  return baseClient.put(`/devices/${id}`, request);
}
