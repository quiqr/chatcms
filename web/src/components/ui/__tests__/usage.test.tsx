import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { HowFull, whyThereIsNoRoom } from "../usage";

describe("how full an organisation is", () => {
  it("says the number and nothing about room when there is no limit", () => {
    render(<HowFull counted={{ held: 18, full: false }} of="seats" />);

    expect(screen.getByText("18 · no limit")).toBeInTheDocument();
    expect(screen.queryByText(/of/)).toBeNull();
  });

  it("says what is used against what is allowed", () => {
    render(<HowFull counted={{ held: 7, allowed: 10, full: false }} of="seats" />);

    expect(screen.getByText("7 of 10 seats used")).toBeInTheDocument();
    expect(screen.queryByText("Almost full")).toBeNull();
  });

  it("warns from four fifths, and says so when it is full", () => {
    const { unmount } = render(<HowFull counted={{ held: 8, allowed: 10, full: false }} of="seats" />);
    expect(screen.getByText("Almost full")).toBeInTheDocument();
    unmount();

    render(<HowFull counted={{ held: 10, allowed: 10, full: true }} of="seats" />);
    expect(screen.getByText("Full")).toBeInTheDocument();
    expect(screen.queryByText("Almost full")).toBeNull();
  });
});

describe("why there is no room", () => {
  it("says nothing while there is room", () => {
    expect(whyThereIsNoRoom({ held: 7, allowed: 10, full: false }, "seats")).toBeNull();
    expect(whyThereIsNoRoom({ held: 99, full: false }, "sites")).toBeNull();
  });

  it("names the limit and what can be done about it", () => {
    const said = whyThereIsNoRoom({ held: 10, allowed: 10, full: true }, "seats");
    expect(said).toContain("All 10 seats are in use");
    // One of something is not "all 1 of them".
    expect(whyThereIsNoRoom({ held: 1, allowed: 1, full: true }, "seats")).toContain(
      "The one seat is in use",
    );
    expect(said).toContain("ask the platform admin");

    const sites = whyThereIsNoRoom({ held: 5, allowed: 5, full: true }, "sites");
    expect(sites).toContain("All 5 sites are in use");
    expect(sites).toContain("Archive or remove one");
  });
});
