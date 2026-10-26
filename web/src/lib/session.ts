import { get, post } from "./api";

export type Person = {
  id: string;
  email: string;
  displayName: string;
};

export function signIn(email: string, password: string, remember: boolean) {
  return post<Person>("/api/sign-in", { email, password, remember });
}

export function signOut() {
  return post<void>("/api/sign-out");
}

export function currentPerson() {
  return get<Person>("/api/me");
}

export function requestPasswordReset(email: string) {
  return post<{ message: string }>("/api/password-reset", { email });
}

export function setPasswordWithToken(token: string, password: string) {
  // Answers with who you now are: setting a password signs you in, because
  // you have just proved everything signing in would ask for.
  return post<Person>("/api/set-password", { token, password });
}
