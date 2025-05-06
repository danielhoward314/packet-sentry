import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Skeleton } from "@/components/ui/skeleton";

type ModeRadioFormProps = {
  field: {
    value: string;
    onChange: (value: string) => void;
  };
  clearError: () => void;
};

export function ModeRadioForm({ field, clearError }: ModeRadioFormProps) {
  return (
    <RadioGroup
      value={field.value}
      onValueChange={(val) => {
        field.onChange(val);
        clearError();
      }}
      className="grid grid-cols-2 gap-16 w-full"
    >
      <div>
        <RadioGroupItem value="light" id="r1" />
        <div className="force-light w-full rounded border p-4 transition-all bg-background text-foreground [&_.skeleton]:bg-muted">
          <Label className="pb-2" htmlFor="r1">
            Light
          </Label>
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10 rounded-full skeleton" />
            <div className="flex flex-col gap-2 flex-1">
              <Skeleton className="h-4 w-3/4 skeleton" />
              <Skeleton className="h-4 w-1/2 skeleton" />
            </div>
          </div>
        </div>
      </div>

      <div>
        <RadioGroupItem value="dark" id="r2" />
        <div className="dark w-full rounded border p-4 transition-all bg-background text-foreground [&_.skeleton]:bg-muted">
          <Label className="pb-2" htmlFor="r2">
            Dark
          </Label>
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10 rounded-full skeleton" />
            <div className="flex flex-col gap-2 flex-1">
              <Skeleton className="h-4 w-3/4 skeleton" />
              <Skeleton className="h-4 w-1/2 skeleton" />
            </div>
          </div>
        </div>
      </div>
    </RadioGroup>
  );
}
