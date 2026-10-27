import { say } from "../lib/say";
export function Placeholder({ screen }: { screen: string }) {
  return (
    <section className="rounded-xl border border-dashed border-input bg-card p-8">
      <h2 className="text-[22px] font-semibold tracking-tight">{screen}</h2>
      <p className="mt-2 text-[15px] text-muted-foreground">
        {say("placeholder.this-part-of-the-site-is")}
      </p>
    </section>
  );
}
