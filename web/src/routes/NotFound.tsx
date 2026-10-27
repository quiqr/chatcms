import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { say } from "../lib/say";

export function NotFound() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-xl flex-col justify-center gap-4 p-6">
      <h1 className="text-[28px] font-semibold tracking-tight">{say("not-found.this-address-does-not-exist")}</h1>
      <p className="text-[15px] text-muted-foreground">
        {say("not-found.the-address-you-followed-does-not")}
      </p>
      <div>
        <Button asChild>
          <NavLink to="/projects">{say("not-found.go-to-your-projects")}</NavLink>
        </Button>
      </div>
    </main>
  );
}
