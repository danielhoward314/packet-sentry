import {
  SignupResponse,
  VerifyResponse,
  LoginResponse,
  ValidateSessionResponse,
  RefreshTokenResponse,
} from "@/types/api";

export function isSignupResponse(obj: unknown): obj is SignupResponse {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof (obj as SignupResponse).token === "string"
  );
}

export function isVerifyResponse(obj: unknown): obj is VerifyResponse {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof (obj as VerifyResponse).adminUiAccessToken === "string" &&
    typeof (obj as VerifyResponse).adminUiRefreshToken === "string" &&
    typeof (obj as VerifyResponse).apiAccessToken === "string" &&
    typeof (obj as VerifyResponse).apiRefreshToken === "string"
  );
}

export function isLoginResponse(obj: unknown): obj is LoginResponse {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof (obj as LoginResponse).administratorId === "string" &&
    typeof (obj as LoginResponse).organizationId === "string" &&
    typeof (obj as LoginResponse).administratorName === "string" &&
    typeof (obj as LoginResponse).organizationName === "string" &&
    typeof (obj as LoginResponse).billingPlan === "string" &&
    typeof (obj as LoginResponse).adminUiAccessToken === "string" &&
    typeof (obj as LoginResponse).adminUiRefreshToken === "string" &&
    typeof (obj as LoginResponse).apiAccessToken === "string" &&
    typeof (obj as LoginResponse).apiRefreshToken === "string"
  );
}

export function isValidateSessionResponse(
  obj: unknown,
): obj is ValidateSessionResponse {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof (obj as ValidateSessionResponse).jwt === "string"
  );
}

export function isRefreshTokenResponse(
  obj: unknown,
): obj is RefreshTokenResponse {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof (obj as RefreshTokenResponse).jwt === "string"
  );
}
