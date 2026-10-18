import { useEffect, useMemo, useState } from "react";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";

import { describeError, listImages, websocketUrl } from "@/lib/api";
import type { ConnectionStatus, FeedImage } from "@/types";
import { insertLive, parseServerEvent, uniqueTags } from "@/lib/feed";
import { useWebSocket } from "./useWebsocket";

export const PAGE_SIZE = 24;

export type FeedStatus = "loading" | "ready" | "error";

export type Feed = {
  images: FeedImage[];
  tags: string[];
  status: FeedStatus;
  error?: string;
  moreError?: string;
  loadingMore: boolean;
  hasMore: boolean;
  connection: ConnectionStatus;
  loadMore: () => void;
  reload: () => void;
  insertLive: (image: FeedImage) => void;
};

export function useFeed(activeTags: string[] = []): Feed {
  const tagsKey = [...activeTags].sort();
  const queryClient = useQueryClient();

  const query = useInfiniteQuery({
    queryKey: ["images", tagsKey],
    queryFn: ({ pageParam, signal }) =>
      listImages({ limit: PAGE_SIZE, cursor: pageParam, tags: tagsKey }, signal),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.nextCursor,
  });

  const handleMessage = (payload: unknown) => {
    const event = parseServerEvent(payload);
    if (event) insertLive(queryClient, event.data);
  };

  const handleReconnect = () => {
    void queryClient.invalidateQueries({ queryKey: ["images"] });
  };

  const connection = useWebSocket({
    url: websocketUrl(),
    onMessage: handleMessage,
    onReconnect: handleReconnect,
  });

  const images = useMemo(() => query.data?.pages.flatMap((page) => page.items) ?? [], [query.data]);
  const [tags, setTags] = useState<string[]>([]);
  useEffect(() => {
    setTags((current) => [...new Set([...current, ...uniqueTags(images)])].sort());
  }, [images]);

  return {
    images,
    tags,
    status: query.data !== undefined ? "ready" : query.isError ? "error" : "loading",
    error: query.error ? describeError(query.error) : undefined,
    moreError:
      query.isFetchNextPageError && !query.isFetchingNextPage
        ? describeError(query.error)
        : undefined,
    loadingMore: query.isFetchingNextPage,
    hasMore: query.hasNextPage,
    connection,
    loadMore: () => void query.fetchNextPage({ cancelRefetch: false }),
    reload: () => void query.refetch(),
    insertLive: (image: FeedImage) => insertLive(queryClient, image),
  };
}
