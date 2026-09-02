import { cn } from "@/lib/utils";

export type StatusTone = "online" | "degraded" | "offline" | "idle" | "unknown";

const TONE_CLASS: Record<StatusTone, string> = {
  online: "text-success",
  degraded: "text-warning",
  offline: "text-destructive",
  idle: "text-muted-foreground",
  unknown: "text-faint-foreground",
};

/**
 * The smallest state carrier on the bench.
 *
 * `pulse` is reserved for a state that is actively changing — a running poll
 * loop, a live stream. A steady state gets a steady dot, so a pulsing dot
 * always means "this is moving right now" rather than "this is green".
 */
export function StatusDot({
  tone = "unknown",
  pulse = false,
  className,
}: {
  tone?: StatusTone;
  pulse?: boolean;
  className?: string;
}) {
  return (
    <span className={cn("relative inline-flex size-2 shrink-0", TONE_CLASS[tone], className)}>
      <span
        className={cn(
          "size-2 rounded-full bg-current",
          pulse && "pulse-dot",
        )}
      />
    </span>
  );
}
