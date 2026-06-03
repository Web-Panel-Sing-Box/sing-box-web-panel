import { apiPost, apiGet } from "./client";
import type { MeResponse } from "./types";

// ------- DTOs -------

export type LoginRequest = {
  username: string;
  password: string;
};

export type LoginResponse = {
  token: string;
};

export type LoginTOTPResponse = {
  requires_totp: true;
  temp_token: string;
};

export type LoginTOTPRequest = {
  temp_token: string;
  code: string;
};

export type LoginRecoveryRequest = {
  username: string;
  code: string;
};

export type SetupTOTPResponse = {
  qr_uri: string;
  secret: string;
};

export type ConfirmTOTPRequest = {
  code: string;
};

export type ConfirmTOTPResponse = {
  message: string;
  recovery_codes: string[];
};

export type DisableTOTPRequest = {
  code: string;
};

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};

export type MessageResponse = {
  message: string;
};

// ------- API functions -------

// Returns a token on success. When 2FA is enabled the backend responds with
// HTTP 403 + { requires_totp, temp_token }, surfaced as an ApiError (see client.ts)
// and handled by the auth context.
export function login(body: LoginRequest): Promise<LoginResponse> {
  return apiPost<LoginResponse>("/auth/login", body);
}

export function loginTOTP(body: LoginTOTPRequest): Promise<LoginResponse> {
  return apiPost<LoginResponse>("/auth/login/totp", body);
}

export function loginRecovery(body: LoginRecoveryRequest): Promise<LoginResponse> {
  return apiPost<LoginResponse>("/auth/login/recovery", body);
}

export function getMe(): Promise<MeResponse> {
  return apiGet<MeResponse>("/auth/me");
}

export function logout(): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/auth/logout");
}

export function setupTOTP(): Promise<SetupTOTPResponse> {
  return apiPost<SetupTOTPResponse>("/auth/totp/setup");
}

export function confirmTOTP(body: ConfirmTOTPRequest): Promise<ConfirmTOTPResponse> {
  return apiPost<ConfirmTOTPResponse>("/auth/totp/confirm", body);
}

export function disableTOTP(body: DisableTOTPRequest): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/auth/totp/disable", body);
}

export function changePassword(body: ChangePasswordRequest): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/auth/change-password", body);
}
