import { NavLink } from "react-router";

import { Button } from "../components/ui/button";

export function NotFound() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center gap-4 p-6">
      <h1 className="text-[28px] font-semibold tracking-tight">This address does not exist</h1>
      <p className="text-[15px] text-muted-foreground">
        The address you followed does not lead anywhere. Nothing is lost; the page you were
        looking for may have been renamed.
      </p>
      <div>
        <Button asChild>
          <NavLink to="/projects">Go to your projects</NavLink>
        </Button>
      </div>
    </main>
  );
}
