import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { Avatar, Tile, initialsOf } from "../avatar";
import { Badge } from "../badge";
import { Button } from "../button";
import { Dialog, DialogContent, DialogTrigger } from "../dialog";
import { Disclosure } from "../disclosure";
import { Menu, MenuContent, MenuItem, MenuTrigger } from "../menu";
import { RadioCards } from "../radio-cards";
import { Segmented } from "../segmented";
import { Table, TableCell, TableHeaderCell, TableWrapper } from "../table";
import { cn } from "../../../lib/cn";

describe("cn", () => {
  it("merges conflicting classes, keeping the last", () => {
    expect(cn("p-2", "p-4")).toBe("p-4");
    expect(cn("text-sm", false && "hidden", "font-medium")).toBe("text-sm font-medium");
  });
});

describe("Button", () => {
  it("renders each variant the designs use", () => {
    for (const variant of ["primary", "ink", "outline", "ghost", "destructive"] as const) {
      const { unmount } = render(<Button variant={variant}>Save</Button>);
      expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
      unmount();
    }
  });

  it("carries a count when it has one", () => {
    render(<Button count={4}>Publish</Button>);
    expect(screen.getByRole("button", { name: /Publish\s*4/ })).toBeInTheDocument();
  });

  it("is a real button, not a styled container", () => {
    render(<Button>Save</Button>);
    expect(screen.getByRole("button").tagName).toBe("BUTTON");
  });
});

describe("Badge", () => {
  it("renders each state", () => {
    for (const tone of ["changed", "new", "published", "live", "attention", "neutral"] as const) {
      const { unmount } = render(<Badge tone={tone}>State</Badge>);
      expect(screen.getByText("State")).toBeInTheDocument();
      unmount();
    }
  });

  it("gives a dot to the states the design draws one on", () => {
    const live = render(<Badge tone="live">Live now</Badge>);
    expect(live.getByTestId("badge-dot")).toBeInTheDocument();
    live.unmount();

    const changed = render(<Badge tone="changed">Changed</Badge>);
    expect(changed.queryByTestId("badge-dot")).toBeNull();
  });
});

describe("Segmented", () => {
  it("reports what was chosen and marks it for assistive technology", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <Segmented
        label="How to show this page"
        value="form"
        onValueChange={onValueChange}
        options={[
          { value: "form", label: "Form" },
          { value: "split", label: "Side by side" },
        ]}
      />,
    );

    const group = screen.getByRole("radiogroup", { name: "How to show this page" });
    expect(within(group).getByRole("radio", { name: "Form" })).toBeChecked();

    await user.click(within(group).getByRole("radio", { name: "Side by side" }));
    expect(onValueChange).toHaveBeenCalledWith("split");
  });
});

describe("RadioCards", () => {
  it("selects one, and is reachable by keyboard", async () => {
    const user = userEvent.setup();

    function Harness() {
      const [value, setValue] = useState("main");
      return (
        <RadioCards
          label="Push to"
          value={value}
          onValueChange={setValue}
          options={[
            { value: "main", title: "Main branch" },
            { value: "separate", title: "A separate branch" },
          ]}
        />
      );
    }

    render(<Harness />);

    expect(screen.getByRole("radio", { name: /Main branch/ })).toBeChecked();

    // Tab reaches the chosen option, which is where arrow keys take over.
    // Roving focus itself belongs to the primitive and is not re-tested here.
    await user.tab();
    expect(screen.getByRole("radio", { name: /Main branch/ })).toHaveFocus();

    await user.click(screen.getByRole("radio", { name: /A separate branch/ }));
    expect(screen.getByRole("radio", { name: /A separate branch/ })).toBeChecked();
    expect(screen.getByRole("radio", { name: /Main branch/ })).not.toBeChecked();
  });
});

describe("Menu", () => {
  it("opens and closes", async () => {
    const user = userEvent.setup();

    render(
      <Menu>
        <MenuTrigger>Actions</MenuTrigger>
        <MenuContent>
          <MenuItem>Rename</MenuItem>
          <MenuItem destructive>Delete...</MenuItem>
        </MenuContent>
      </Menu>,
    );

    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(await screen.findByRole("menuitem", { name: "Rename" })).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(screen.queryByRole("menuitem", { name: "Rename" })).toBeNull();
  });
});

describe("Dialog", () => {
  it("opens, names itself and closes", async () => {
    const user = userEvent.setup();

    render(
      <Dialog>
        <DialogTrigger>Go back to this version...</DialogTrigger>
        <DialogContent title="Go back to this version?" description="Nothing is lost.">
          <p>Body</p>
        </DialogContent>
      </Dialog>,
    );

    await user.click(screen.getByRole("button", { name: /Go back to this version/ }));

    const dialog = await screen.findByRole("dialog", { name: "Go back to this version?" });
    expect(within(dialog).getByText("Nothing is lost.")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).toBeNull();
  });
});

describe("Disclosure", () => {
  it("hides its content until it is opened", async () => {
    const user = userEvent.setup();

    render(
      <Disclosure summary="Technical details">
        <p>The inner workings</p>
      </Disclosure>,
    );

    expect(screen.queryByText("The inner workings")).toBeNull();
    await user.click(screen.getByRole("button", { name: "Technical details" }));
    expect(screen.getByText("The inner workings")).toBeInTheDocument();
  });
});

describe("Table", () => {
  it("scrolls on its own rather than widening the page", () => {
    render(
      <TableWrapper data-testid="wrapper">
        <Table>
          <thead>
            <tr>
              <TableHeaderCell>Person</TableHeaderCell>
            </tr>
          </thead>
          <tbody>
            <tr>
              <TableCell>Joris</TableCell>
            </tr>
          </tbody>
        </Table>
      </TableWrapper>,
    );

    expect(screen.getByTestId("wrapper").className).toContain("overflow-x-auto");
    expect(screen.getByRole("table")).toBeInTheDocument();
  });
});

describe("Avatar and Tile", () => {
  it("shortens a name the way the designs do", () => {
    expect(initialsOf("Joris Bakker")).toBe("JB");
    expect(initialsOf("Marieke")).toBe("MA");
    expect(initialsOf("  ")).toBe("?");
  });

  it("names the person it stands for", () => {
    render(<Avatar name="Joris Bakker" />);
    expect(screen.getByRole("img", { name: "Joris Bakker" })).toBeInTheDocument();
  });

  it("marks a pending invitation", () => {
    render(<Avatar name="Sam Visser" pending />);
    expect(screen.getByRole("img", { name: "Sam Visser" }).className).toContain("border-dashed");
  });

  it("renders a project tile without announcing it twice", () => {
    render(<Tile name="Korenaar Bakery" />);
    expect(screen.getByText("K")).toHaveAttribute("aria-hidden", "true");
  });
});
