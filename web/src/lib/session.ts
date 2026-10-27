import { get, post, put, remove, sendForm } from "./api";

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

export function changeOwnName(displayName: string) {
  return put<Person>("/api/account/name", { displayName });
}

export function changeOwnPassword(current: string, next: string) {
  return put<void>("/api/account/password", { current, next });
}

export function signOutEverywhere() {
  return post<void>("/api/sign-out/everywhere");
}

export type WhereSignedIn = {
  id: string;
  lastSeen: string;
  started: string;
  thisOne: boolean;
};

export function setPicture(file: File) {
  const carried = new FormData();
  carried.append("picture", file);
  return sendForm<void>("/api/account/picture", carried, "PUT");
}

export function removePicture() {
  return remove<void>("/api/account/picture");
}

export function askForANewAddress(email: string) {
  return post<void>("/api/account/address", { email });
}

export function whereSignedIn() {
  return get<{ sessions: WhereSignedIn[] }>("/api/account/sessions");
}

export function endSession(id: string) {
  return remove<void>(`/api/account/sessions/${id}`);
}

export function proveTheNewAddress(token: string) {
  return post<Person>("/api/confirm-address", { token });
}

export function chooseAppearance(appearance: string) {
  return put<void>("/api/account/appearance", { appearance });
}
