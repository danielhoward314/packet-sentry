export interface SignupRequest {
  organizationName: string;
  primaryAdministratorEmail: string;
  primaryAdministratorName: string;
  primaryAdministratorCleartextPassword: string;
}

export interface SignupResponse {
  token: string;
}

export interface VerifyRequest {
  token: string;
  verificationCode: string;
}

export interface VerifyResponse {
  adminUiAccessToken: string;
  adminUiRefreshToken: string;
  apiAccessToken: string;
  apiRefreshToken: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  administratorId: string;
  organizationId: string;
  administratorName: string;
  organizationName: string;
  billingPlan: string;
  adminUiAccessToken: string;
  adminUiRefreshToken: string;
  apiAccessToken: string;
  apiRefreshToken: string;
}

export interface ValidateSessionRequest {
  jwt: string;
}

export interface ValidateSessionResponse {
  jwt: string;
}

export enum ClaimsType {
  UNSPECIFIED = 0,
  ADMIN_UI_SESSION = 1,
  API_AUTHORIZATION = 2,
}

export interface RefreshTokenRequest {
  jwt: string;
  claimsType: ClaimsType;
}

export interface RefreshTokenResponse {
  jwt: string;
}

export interface ResetVerifyRequest {
  email: string;
}

export enum CredentialType {
  PASSWORD = 0,
  EMAIL_VERIFICATION_CODE = 1,
}

export enum IdentifierType {
  ID = 0,
  EMAIL = 1,
}

export interface ResetPasswordRequest {
  credential: string;
  credentialType: CredentialType;
  identifier: string;
  identifierType: IdentifierType;
  newPassword: string;
  confirmNewPassword: string;
}

export interface CreateAdministratorRequest {
  organizationId: string;
  email: string;
  displayName: string;
  authorizationRole: string;
}

export interface ActivateAdministratorRequest {
  token: string;
  verificationCode: string;
  password: string;
}

export type GetAdministratorResponse = {
  id: string;
  email: string;
  displayName: string;
  organizationId: string;
  verified: boolean;
  authorizationRole: "PRIMARY_ADMIN" | "SECONDARY_ADMIN";
};

export interface UpdateAdministratorRequest {
  email: string;
  displayName: string;
  authorizationRole: string;
}

export type GetOrganizationResponse = {
  id: string;
  organizationName: string;
  billingPlan: string;
};

export type UpdateOrganizationRequest = {
  name: string;
  billingPlan: string;
};
