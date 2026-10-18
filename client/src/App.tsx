import { EmptyState } from "@/components/empty-state";
import { ImageCard } from "@/components/image-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useFeed } from "@/hooks/useFeed";
import { useCallback, useState } from "react";
import { TagFilter } from "./components/tag-filter";
import { Upload } from "lucide-react";
import { UploadDialog } from "./components/upload-dialog";
import { useInfiniteScroll } from "./hooks/useInfiniteScroll";
import { Dialog, DialogContent, DialogTitle } from "./components/ui/dialog";
import type { FeedImage } from "./types";

export function App() {
  const [activeTags, setActiveTags] = useState<string[]>([]);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [preview, setPreview] = useState<FeedImage | null>(null);
  const feed = useFeed(activeTags);
  const sentinelRef = useInfiniteScroll(feed.loadMore, feed.status === "ready" && feed.hasMore);

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
          <>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {feed.images.map((image) => (
                <ImageCard
                  key={image.id}
                  image={image}
                  activeTags={activeTags}
                  onToggleTag={toggleTag}
                  onOpen={setPreview}
                />
              ))}
            </div>

            <div ref={sentinelRef} aria-hidden className="h-px" />

            {feed.loadingMore && <FeedSkeleton count={4} />}

            {feed.moreError && (
              <div className="flex flex-col items-center gap-2 py-4">
                <p className="text-sm text-destructive">{feed.moreError}</p>
                <Button variant="outline" size="sm" onClick={feed.loadMore}>
                  Load more
                </Button>
              </div>
            )}

            {!feed.hasMore && !feed.loadingMore && (
              <p className="py-4 text-center text-xs text-muted-foreground">
                That is everything so far.
              </p>
            )}
          </>
        )}
      </main>

      <UploadDialog open={uploadOpen} onOpenChange={setUploadOpen} onUploaded={feed.insertLive} />

      <Dialog open={preview !== null} onOpenChange={(open) => !open && setPreview(null)}>
        <DialogContent
          showCloseButton={false}
          aria-describedby={undefined}
          className="max-w-[min(92vw,64rem)] border-0 bg-transparent p-0 shadow-none"
        >
          <DialogTitle className="sr-only">{preview?.title}</DialogTitle>
          {preview && (
            <img
              src={preview.url}
              alt={preview.title}
              className="max-h-[85dvh] w-full rounded-xl object-contain"
            />
          )}
        </DialogContent>
      </Dialog>
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
