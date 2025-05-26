import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { formatInTimeZone } from "date-fns-tz";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Helper to format Date as UTC string "yyyy-MM-dd HH:mm"
export function formatUTC(date: Date): string {
  return formatInTimeZone(date, "UTC", "yyyy-MM-dd HH:mm");
}