import { UpdateAdministratorRequest } from '@/types/api'
import baseClient from './axiosBaseClient'
import { AxiosResponse } from 'axios'

export async function getOrganization(id: string): Promise<any> {
  const res = await baseClient.get(`/organizations/${id}`)
  return res.data
}

export async function updateAdministrators(
    id: string,
    request: UpdateAdministratorRequest
  ): Promise<AxiosResponse<void>> {
    return baseClient.put(`/administrators/${id}`, request)
}

export async function listAdministrators(organizationId: string): Promise<any> {
  const res = await baseClient.get(`/administrators?organizationId=${organizationId}`)
  return res.data
}