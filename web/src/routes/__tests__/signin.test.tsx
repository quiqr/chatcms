import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SignInPage } from "../SignInPage";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/sign-in"]}>
      <SignInPage />
    </MemoryRouter>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("the sign-in screen", () => {
  it("renders the form the designs draw", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Sign in" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByLabelText("Keep me signed in on this device")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Forgot password?" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
  });

  it("tells someone with an invitation where to start", () => {
    renderPage();

    expect(
      screen.getByText(/Got an invitation email\? Use the link in it to set your password first\./),
    ).toBeInTheDocument();
  });

  it("carries the line from the designs", () => {
    renderPage();

    expect(screen.getByText(/Fill in a form, or just ask\./)).toBeInTheDocument();
  });

  it("sends the request token with the sign-in", async () => {
    document.cookie = "chatcms_request_token=a-token; path=/";

    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: "1", email: "joris@example.org", displayName: "Joris" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    renderPage();

    await user.type(screen.getByLabelText("Email"), "joris@example.org");
    await user.type(screen.getByLabelText("Password"), "a long enough passphrase");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/sign-in",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ "X-Chatcms-Request-Token": "a-token" }),
      }),
    );
  });

  it("shows what the server said when the sign-in is refused", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({
          error: { code: "unauthorized", message: "That email address and password do not match." },
        }),
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.type(screen.getByLabelText("Email"), "joris@example.org");
    await user.type(screen.getByLabelText("Password"), "wrong one entirely");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "That email address and password do not match.",
    );
  });

  it("says something useful when the server cannot be reached", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network down")));

    const user = userEvent.setup();
    renderPage();

    await user.type(screen.getByLabelText("Email"), "joris@example.org");
    await user.type(screen.getByLabelText("Password"), "a long enough passphrase");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Nothing is lost");
  });
});
