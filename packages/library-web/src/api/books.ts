import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  AddBookToShelfRequestSchema,
  BookService,
  CreateShelfRequestSchema,
  DeleteBookRequestSchema,
  EmptyBookTrashRequestSchema,
  GetBookRequestSchema,
  GetReadingProgressRequestSchema,
  ImportBookRequestSchema,
  ListBooksRequestSchema,
  ListShelvesRequestSchema,
  PurgeBookRequestSchema,
  RemoveBookFromShelfRequestSchema,
  RestoreBookRequestSchema,
  UpdateBookRequestSchema,
  UpdateReadingProgressRequestSchema,
} from "@realtime-me/library-contracts";
import type {
  Book,
  BookFormat,
  ReadingProgress,
  Shelf,
} from "@realtime-me/library-contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export type BookListOptions = Readonly<{
  query?: string;
  shelfUid?: string;
  format?: BookFormat;
  trashed?: boolean;
  pageSize?: number;
  pageToken?: string;
}>;

export type BookListPage = Readonly<{
  books: Book[];
  nextPageToken: string;
}>;

export class BooksClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof BookService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(BookService, privateTransport(baseUrl));
  }
  async listPage(
    options: BookListOptions = {},
    signal?: AbortSignal,
  ): Promise<BookListPage> {
    const response = await this.client.listBooks(
      create(ListBooksRequestSchema, options),
      signal ? { signal } : undefined,
    );
    return {
      books: response.books,
      nextPageToken: response.nextPageToken,
    };
  }
  async get(bookUid: string): Promise<Book> {
    return required(
      (await this.client.getBook(create(GetBookRequestSchema, { bookUid })))
        .book,
      "book",
    );
  }
  async importUpload(uploadUid: string): Promise<Book> {
    return required(
      (
        await this.client.importBook(
          create(ImportBookRequestSchema, { uploadUid }),
        )
      ).book,
      "book",
    );
  }
  async update(book: Book): Promise<Book> {
    return required(
      (
        await this.client.updateBook(
          create(UpdateBookRequestSchema, {
            bookUid: book.uid,
            title: book.title,
            authors: book.authors,
            series: book.series,
            seriesNumber: book.seriesNumber,
            description: book.description,
          }),
        )
      ).book,
      "book",
    );
  }
  async trash(bookUid: string): Promise<void> {
    await this.client.deleteBook(create(DeleteBookRequestSchema, { bookUid }));
  }
  async restore(bookUid: string): Promise<void> {
    await this.client.restoreBook(
      create(RestoreBookRequestSchema, { bookUid }),
    );
  }
  async purge(bookUid: string): Promise<void> {
    await this.client.purgeBook(create(PurgeBookRequestSchema, { bookUid }));
  }
  async emptyTrash(): Promise<void> {
    await this.client.emptyBookTrash(create(EmptyBookTrashRequestSchema));
  }
  async progress(bookUid: string): Promise<ReadingProgress> {
    return required(
      (
        await this.client.getReadingProgress(
          create(GetReadingProgressRequestSchema, { bookUid }),
        )
      ).readingProgress,
      "readingProgress",
    );
  }
  async saveProgress(readingProgress: ReadingProgress): Promise<void> {
    await this.client.updateReadingProgress(
      create(UpdateReadingProgressRequestSchema, { readingProgress }),
    );
  }
  async shelves(signal?: AbortSignal): Promise<Shelf[]> {
    return (
      await this.client.listShelves(
        create(ListShelvesRequestSchema),
        signal ? { signal } : undefined,
      )
    ).shelves;
  }
  async createShelf(displayName: string): Promise<Shelf> {
    return required(
      (
        await this.client.createShelf(
          create(CreateShelfRequestSchema, { displayName }),
        )
      ).shelf,
      "shelf",
    );
  }
  async addToShelf(shelfUid: string, bookUid: string): Promise<void> {
    await this.client.addBookToShelf(
      create(AddBookToShelfRequestSchema, { shelfUid, bookUid }),
    );
  }
  async removeFromShelf(shelfUid: string, bookUid: string): Promise<void> {
    await this.client.removeBookFromShelf(
      create(RemoveBookFromShelfRequestSchema, { shelfUid, bookUid }),
    );
  }
  contentUrl(book: Book): string {
    return resolveApiUrl(this.baseUrl, book.contentUrl);
  }
  coverUrl(book: Book): string {
    return resolveApiUrl(this.baseUrl, book.coverUrl);
  }
}
