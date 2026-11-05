import { Button } from "../components/ui/button";
import { say, type Key } from "../lib/say";

const passphrase = "a long enough passphrase";

const people: { email: string; role: Key }[] = [
  { email: "admin@example.org", role: "platform.platform-admin" },
  { email: "owner@example.org", role: "role.implementer" },
  { email: "marieke@korenaar.example", role: "role.editor" },
  { email: "sam@korenaar.example", role: "role.viewer" },
];

export default function DevSignIn({
  busy,
  onPick,
}: {
  busy: boolean;
  onPick: (email: string, password: string) => void;
}) {
  return (
    <section
      aria-labelledby="dev-sign-in"
      className="mt-6 rounded-md border border-dashed border-input p-4"
    >
      <h2 id="dev-sign-in" className="text-[13px] font-semibold text-muted-foreground">
        {say("dev-sign-in.development-only")}
      </h2>
      <p className="mt-1 text-[13px] text-subtle-foreground">{say("dev-sign-in.sign-in-as-a-seeded")}</p>
      <div className="mt-3 grid grid-cols-2 gap-2">
        {people.map((person) => (
          <Button
            key={person.email}
            type="button"
            variant="outline"
            size="secondary"
            disabled={busy}
            title={person.email}
            onClick={() => onPick(person.email, passphrase)}
          >
            {say(person.role)}
          </Button>
        ))}
      </div>
    </section>
  );
}
