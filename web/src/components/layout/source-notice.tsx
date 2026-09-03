"use client";

import { PlugZap, TriangleAlert } from "lucide-react";

import { useDictionary } from "@/components/locale-provider";
import { CopyButton } from "@/components/ui/copy-button";
import { useTelemetry } from "@/lib/telemetry/use-telemetry";

/**
 * What to do when the live source has nothing to show.
 *
 * A console that simply renders zeros when the API is down is the worst
 * possible outcome: every panel looks like a healthy, idle gateway. So an
 * unreachable link takes over the top of the page and says which process is
 * missing and how to start it — the two things the reader needs and the two
 * things a spinner never tells them.
 *
 * A gateway that *is* answering but reporting an error gets a quieter band:
 * the readings below it are real and still worth looking at.
 */

/**
 * One command per line, each in its own terminal.
 *
 * Not chained with a shell operator: the three processes are long-running, and
 * the operator on this bench may be in PowerShell, where the POSIX way of
 * backgrounding them is a syntax error.
 */
const START_COMMAND = `go run ./cmd/ft12-emulator
go run ./cmd/ft12-gateway
go run ./cmd/ft12-api`;

export function SourceNotice() {
  const dict = useDictionary();
  const { source, link, error } = useTelemetry();

  if (source !== "live") return null;

  if (link === "unreachable") {
    return (
      <section
        role="status"
        className="mb-4 rounded-md border border-destructive/40 bg-destructive/8 p-3.5"
      >
        <div className="flex items-start gap-2.5">
          <PlugZap className="mt-0.5 size-4 shrink-0 text-destructive" />
          <div className="min-w-0 space-y-2">
            <h2 className="text-[0.8125rem] font-semibold text-foreground">
              {dict.source.unreachableTitle}
            </h2>
            <p className="text-xs leading-relaxed text-muted-foreground">
              {dict.source.unreachableBody}
            </p>
            <p className="text-xs text-faint-foreground">{dict.source.unreachableCommand}</p>
            <div className="flex items-center gap-2">
              <code className="min-w-0 flex-1 overflow-x-auto rounded border border-border bg-surface-sunken px-2 py-1.5 font-mono text-[0.6875rem] leading-relaxed whitespace-pre text-foreground">
                {START_COMMAND}
              </code>
              <CopyButton value={START_COMMAND} />
            </div>
            {error ? (
              <p className="font-mono text-[0.6875rem] text-faint-foreground">{error}</p>
            ) : null}
          </div>
        </div>
      </section>
    );
  }

  if (!error) return null;

  return (
    <section
      role="status"
      className="mb-4 flex items-start gap-2.5 rounded-md border border-warning/40 bg-warning/8 px-3.5 py-2.5"
    >
      <TriangleAlert className="mt-0.5 size-4 shrink-0 text-warning" />
      <div className="min-w-0">
        <p className="text-xs font-medium text-foreground">{dict.source.upstreamError}</p>
        <p className="mt-0.5 font-mono text-[0.6875rem] break-words text-muted-foreground">
          {error}
        </p>
      </div>
    </section>
  );
}
