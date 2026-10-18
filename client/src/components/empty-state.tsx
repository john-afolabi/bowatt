import type { ReactNode } from "react";

import { Card, CardContent } from "@/components/ui/card";

type Props = {
  title: string;
  description: string;
  action?: ReactNode;
};

export function EmptyState({ title, description, action }: Props) {
  return (
    <Card className="border-dashed bg-transparent shadow-none">
      <CardContent className="flex flex-col items-center gap-3 px-6 py-14 text-center">
        <h2 className="text-base font-medium">{title}</h2>
        <p className="max-w-md text-sm text-muted-foreground">{description}</p>
        {action}
      </CardContent>
    </Card>
  );
}
