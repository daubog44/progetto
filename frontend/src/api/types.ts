
// src/api/types.ts
export interface RegisterPayload {
  email: string;
  password: string;
  username: string;
}
export interface RegisterResponse {
  $schema: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user_id: string;
}


export interface LoginPayload {
  email: string;   
  password: string;
}
export type LoginResponse = RegisterResponse;

export interface ApiError {
  error?: string;
  message?: string;
  errors?: Array<{ field?: string; message?: string }>;
}
