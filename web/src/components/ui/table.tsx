import type { HTMLAttributes, TableHTMLAttributes } from "react";

import { cn } from "../../lib/cn";

export function TableWrapper({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("w-full overflow-x-auto", className)} {...props}>
      {children}
    </div>
  );
}

export function Table({ className, ...props }: TableHTMLAttributes<HTMLTableElement>) {
  return <table className={cn("w-full border-collapse text-left text-sm", className)} {...props} />;
}

export function TableHeaderCell({ className, ...props }: HTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      scope="col"
      className={cn("border-b border-border px-3 py-2 text-xs font-medium text-subtle-foreground", className)}
      {...props}
    />
  );
}

export function TableCell({ className, ...props }: HTMLAttributes<HTMLTableCellElement>) {
  return <td className={cn("border-b border-border-subtle px-3 py-3 align-middle", className)} {...props} />;
}
