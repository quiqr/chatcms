import { Slot, Slottable } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type { ButtonHTMLAttributes } from "react";

import { cn } from "../../lib/cn";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md font-semibold whitespace-nowrap transition-colors disabled:pointer-events-none disabled:opacity-50 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        primary: "bg-primary text-primary-foreground hover:bg-primary-hover",
        ink: "bg-secondary text-secondary-foreground hover:opacity-90",
        outline: "border border-input bg-card text-foreground hover:bg-muted",
        ghost: "text-muted-foreground hover:bg-muted hover:text-foreground",
        destructive: "bg-destructive text-primary-foreground hover:opacity-90",
      },
      size: {
        default: "h-11 px-4 text-[15px]",
        compact: "h-10 px-3 text-[15px]",
        secondary: "h-9 px-3 text-sm",
        icon: "size-11",
        iconCompact: "size-9",
      },
    },
    defaultVariants: { variant: "primary", size: "default" },
  },
);

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
    count?: number;
  };

export function Button({
  className,
  variant,
  size,
  asChild = false,
  count,
  children,
  ...props
}: ButtonProps) {
  const Component = asChild ? Slot : "button";

  return (
    <Component className={cn(buttonVariants({ variant, size }), className)} {...props}>
      <Slottable>{children}</Slottable>
      {count !== undefined && (
        <span className="rounded-full bg-white/20 px-2 py-0.5 text-xs font-semibold tabular-nums">
          {count}
        </span>
      )}
    </Component>
  );
}

export { buttonVariants };
