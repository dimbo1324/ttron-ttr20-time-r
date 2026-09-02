import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Merges class names, resolving Tailwind conflicts by last-wins.
 *
 * `clsx` handles conditionals; `twMerge` handles the conflicts `clsx` cannot see.
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
