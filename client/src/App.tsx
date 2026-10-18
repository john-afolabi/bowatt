import { EmptyState } from "@/components/empty-state";
import { ImageCard } from "@/components/image-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useFeed } from "@/hooks/useFeed";

export function App() {
  const feed = useFeed();

  return (
    <div className="min-h-dvh bg-background">
      <header className="sticky top-0 z-10 border-b bg-background/80 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center gap-2 px-4 py-3 sm:gap-3 sm:px-6">
          <span className="mr-auto text-lg font-semibold tracking-tight">Bowatt</span>
        </div>
      </header>

      <main className="mx-auto max-w-7xl space-y-5 px-4 py-6 sm:px-6">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">Live feed</h1>
          <p className="text-sm text-muted-foreground">Anonymous uploads, newest first.</p>
        </div>

        {feed.status === "loading" && <FeedSkeleton />}

        {feed.status === "error" && (
          <EmptyState
            title="The feed could not be loaded"
            description={feed.error ?? "The server did not respond. Try again in a moment."}
            action={
              <Button variant="outline" onClick={feed.reload}>
                Try again
              </Button>
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
              <ImageCard key={image.id} image={image} />
            ))}
          </div>
        )}
      </main>
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
