export function SignInPage() {
  return (
    <main className="flex min-h-dvh flex-wrap">
      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center p-8">
        <div className="w-full max-w-[380px]">
          <img src="/quiqr-logo.svg" alt="Quiqr ChatCMS" className="h-7 w-auto" />
          <h1 className="mt-8 text-[28px] font-semibold tracking-tight">Sign in</h1>
          <p className="mt-2 text-[15px] text-muted-foreground">
            Got an invitation email? Use the link in it to set your password first.
          </p>
        </div>
      </div>

      <div className="flex min-w-0 shrink grow basis-[420px] items-center justify-center bg-primary-soft p-8">
        <p className="max-w-[320px] text-[15px] text-primary-soft-foreground">
          Fill in a form, or just ask. Your site updates as you go.
        </p>
      </div>
    </main>
  );
}
