export function Placeholder({ screen }: { screen: string }) {
  return (
    <section className="rounded-xl border border-dashed border-input bg-card p-8">
      <h2 className="text-[22px] font-semibold tracking-tight">{screen}</h2>
      <p className="mt-2 text-[15px] text-muted-foreground">
        This part of the site is being built. The frame around it is finished, so what belongs
        here can be added without moving anything else.
      </p>
    </section>
  );
}
