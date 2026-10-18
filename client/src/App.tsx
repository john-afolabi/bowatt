import { EmptyState } from "@/components/empty-state";
import { ImageCard } from "@/components/image-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useFeed } from "@/hooks/useFeed";
import { useCallback, useState } from "react";
import { TagFilter } from "./components/tag-filter";
import { Upload } from "lucide-react";
import { UploadDialog } from "./components/upload-dialog";

export function App() {
  const [activeTags, setActiveTags] = useState<string[]>([]);
  const [uploadOpen, setUploadOpen] = useState(false);
  const feed = useFeed(activeTags);

  const toggleTag = useCallback((tag: string) => {
    setActiveTags((current) =>
      current.includes(tag) ? current.filter((entry) => entry !== tag) : [...current, tag],
    );
  }, []);
  const clearTags = useCallback(() => setActiveTags([]), []);

  return (
    <div className="min-h-dvh bg-background">
      <header className="sticky top-0 z-10 border-b bg-background/80 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center gap-2 px-4 py-3 sm:gap-3 sm:px-6">
          <span className="mr-auto text-lg font-semibold tracking-tight">Bowatt</span>
          <Button onClick={() => setUploadOpen(true)}>
            <Upload aria-hidden />
            Upload
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-7xl space-y-5 px-4 py-6 sm:px-6">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">Live feed</h1>
          <p className="text-sm text-muted-foreground">Anonymous uploads, newest first.</p>
        </div>

        {feed.status !== "error" && (
          <TagFilter
            tags={feed.tags}
            activeTags={activeTags}
            onToggle={toggleTag}
            onClear={clearTags}
          />
        )}

        {feed.status === "loading" && <FeedSkeleton />}

        {feed.status === "error" && (
          <EmptyState
            title={activeTags.length > 0 ? "Nothing matches those tags" : "No images yet"}
            description={
              activeTags.length > 0
                ? "Try removing a tag, or upload the first image with it."
                : "Be the first: upload an image and it will appear here for everyone watching."
            }
            action={
              activeTags.length > 0 ? (
                <Button variant="outline" onClick={clearTags}>
                  Clear filters
                </Button>
              ) : undefined
            }
          />
        )}

        {feed.status === "ready" && feed.images.length === 0 && (
          <EmptyState
            title="No images yet"
            description="Be the first: upload an image and it will appear here for everyone"
          />
        )}

        {feed.status === "ready" && feed.images.length > 0 && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {feed.images.map((image) => (
              <ImageCard
                key={image.id}
                image={image}
                activeTags={activeTags}
                onToggleTag={toggleTag}
              />
            ))}
          </div>
        )}
      </main>

      <UploadDialog
        open={uploadOpen}
        onOpenChange={setUploadOpen}
        onUploaded={() => console.log("uploaded")}
      />
    </div>
  );
}

function FeedSkeleton({ count = 8 }: { count?: number }) {
  return (
    <div
      className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
      aria-hidden
    >
      {Array.from({ length: count }, (_, index) => (
        <div key={index} className="space-y-3 rounded-xl border bg-card p-3">
          <Skeleton className="aspect-4/3 w-full" />
          <Skeleton className="h-4 w-2/3" />
          <Skeleton className="h-5 w-1/3" />
        </div>
      ))}
    </div>
  );
}
