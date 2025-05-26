import { useState } from "react";
import { subMinutes, subHours, subDays, parse } from "date-fns";
import { CalendarIcon } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatUTC } from "@/lib/utils";

interface DateTimeRangePickerProps {
  onChange: (range: { start: Date | null; end: Date | null }) => void;
}

const presets = [
  {
    label: "Last 15 minutes",
    range: () => [subMinutes(new Date(), 15), new Date()],
  },
  { label: "Last 1 hour", range: () => [subHours(new Date(), 1), new Date()] },
  {
    label: "Last 24 hours",
    range: () => [subHours(new Date(), 24), new Date()],
  },
  { label: "Last 7 days", range: () => [subDays(new Date(), 7), new Date()] },
];

// Helper to parse input string as UTC time
function parseUTC(value: string): Date | null {
  const parsedDate = parse(value, "yyyy-MM-dd HH:mm", new Date());
  if (isNaN(parsedDate.getTime())) return null;

  // Adjust parsed local time to UTC by subtracting the timezone offset
  const utcTimestamp = parsedDate.getTime() - parsedDate.getTimezoneOffset() * 60 * 1000;
  return new Date(utcTimestamp);
}

export default function DateTimeRangePicker({
  onChange,
}: DateTimeRangePickerProps) {
  const [startDate, setStartDate] = useState<Date | null>(null);
  const [endDate, setEndDate] = useState<Date | null>(null);
  const [startInput, setStartInput] = useState<string>("");
  const [endInput, setEndInput] = useState<string>("");

  const handleInputChange = (value: string, isStart: boolean) => {
    const parsed = parseUTC(value);
    if (parsed) {
      if (isStart) {
        setStartDate(parsed);
        setStartInput(value);
      } else {
        setEndDate(parsed);
        setEndInput(value);
      }
      onChange({
        start: isStart ? parsed : startDate,
        end: isStart ? endDate : parsed,
      });
    } else {
      if (isStart) setStartInput(value);
      else setEndInput(value);
    }
  };

  const handleCalendarSelect = (date: Date, isStart: boolean) => {
    // The calendar returns a Date object in local time,
    // so convert that to a UTC Date by subtracting the timezone offset
    const utcTimestamp = date.getTime() - date.getTimezoneOffset() * 60 * 1000;
    const utcDate = new Date(utcTimestamp);

    if (isStart) {
      setStartDate(utcDate);
      setStartInput(formatUTC(utcDate));
      onChange({ start: utcDate, end: endDate });
    } else {
      setEndDate(utcDate);
      setEndInput(formatUTC(utcDate));
      onChange({ start: startDate, end: utcDate });
    }
  };

  const applyPreset = (start: Date, end: Date) => {
    // Presets use new Date() which is local time, so convert to UTC
    const startUTC = new Date(start.getTime() - start.getTimezoneOffset() * 60 * 1000);
    const endUTC = new Date(end.getTime() - end.getTimezoneOffset() * 60 * 1000);

    setStartDate(startUTC);
    setEndDate(endUTC);
    setStartInput(formatUTC(startUTC));
    setEndInput(formatUTC(endUTC));
    onChange({ start: startUTC, end: endUTC });
  };

  return (
    <div className="grid gap-4">
      <div className="flex justify-end">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">Presets</Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            {presets.map((preset) => (
              <DropdownMenuItem
                key={preset.label}
                onClick={() => {
                  const [start, end] = preset.range();
                  applyPreset(start, end);
                }}
              >
                {preset.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2">
        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium">Start</label>
          <div className="flex items-center gap-2">
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline" size="icon">
                  <CalendarIcon className="w-4 h-4" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0">
                <Calendar
                  mode="single"
                  selected={startDate ?? undefined}
                  onSelect={(date: Date | undefined) =>
                    date && handleCalendarSelect(date, true)
                  }
                  autoFocus
                />
              </PopoverContent>
            </Popover>
            <Input
              value={startInput}
              onChange={(e) => handleInputChange(e.target.value, true)}
              placeholder="YYYY-MM-DD HH:MM"
            />
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium">End</label>
          <div className="flex items-center gap-2">
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline" size="icon">
                  <CalendarIcon className="w-4 h-4" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0">
                <Calendar
                  mode="single"
                  selected={endDate ?? undefined}
                  onSelect={(date) => date && handleCalendarSelect(date, false)}
                  autoFocus
                />
              </PopoverContent>
            </Popover>
            <Input
              value={endInput}
              onChange={(e) => handleInputChange(e.target.value, false)}
              placeholder="YYYY-MM-DD HH:MM"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
