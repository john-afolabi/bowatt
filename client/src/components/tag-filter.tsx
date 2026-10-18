import { Badge } from "@/components/ui/badge";

type Props = {
  tags: string[];
  activeTags: string[];
  onToggle: (tag: string) => void;
  onClear: () => void;
};

export function TagFilter({ tags, activeTags, onToggle, onClear }: Props) {
  if (tags.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-2" role="group" aria-label="Filter by tag">
      <Badge
        asChild
        variant={activeTags.length === 0 ? "default" : "outline"}
        className="cursor-pointer"
      >
        <button type="button" onClick={onClear} aria-pressed={activeTags.length === 0}>
          All
        </button>
      </Badge>

      {tags.map((tag) => {
        const active = activeTags.includes(tag);
        return (
          <Badge
            key={tag}
            asChild
            variant={active ? "default" : "outline"}
            className="cursor-pointer"
          >
            <button type="button" onClick={() => onToggle(tag)} aria-pressed={active}>
              {tag}
            </button>
          </Badge>
        );
      })}
    </div>
  );
}
