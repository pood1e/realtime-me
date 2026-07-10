package transport

import (
	"context"

	"connectrpc.com/connect"
	booksv1 "example.com/cloud-drive/api/gen/cloud/books/v1"
	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/domain"
)

type bookServer struct{ service *app.BookService }

func (s *bookServer) GetBook(ctx context.Context, request *connect.Request[booksv1.GetBookRequest]) (*connect.Response[booksv1.GetBookResponse], error) {
	book, err := s.service.Get(ctx, request.Msg.GetBookUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.GetBookResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) ListBooks(ctx context.Context, request *connect.Request[booksv1.ListBooksRequest]) (*connect.Response[booksv1.ListBooksResponse], error) {
	page, err := s.service.List(ctx, request.Msg.GetQuery(), request.Msg.GetShelfUid(), bookFormatDomain(request.Msg.GetFormat()), request.Msg.GetTrashed(), int(request.Msg.GetPageSize()), request.Msg.GetPageToken())
	if err != nil {
		return nil, connectError(err)
	}
	books := make([]*booksv1.Book, 0, len(page.Books))
	for _, book := range page.Books {
		books = append(books, bookProto(book))
	}
	return connect.NewResponse(&booksv1.ListBooksResponse{Books: books, NextPageToken: page.NextPageToken}), nil
}

func (s *bookServer) ImportBook(ctx context.Context, request *connect.Request[booksv1.ImportBookRequest]) (*connect.Response[booksv1.ImportBookResponse], error) {
	book, err := s.service.Import(ctx, request.Msg.GetUploadUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.ImportBookResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) UpdateBook(ctx context.Context, request *connect.Request[booksv1.UpdateBookRequest]) (*connect.Response[booksv1.UpdateBookResponse], error) {
	book, err := s.service.Update(ctx, domain.Book{UID: request.Msg.GetBookUid(), Title: request.Msg.GetTitle(), Authors: request.Msg.GetAuthors(), Series: request.Msg.GetSeries(), SeriesNumber: request.Msg.GetSeriesNumber(), Description: request.Msg.GetDescription()})
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.UpdateBookResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) DeleteBook(ctx context.Context, request *connect.Request[booksv1.DeleteBookRequest]) (*connect.Response[booksv1.DeleteBookResponse], error) {
	book, err := s.service.Trash(ctx, request.Msg.GetBookUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.DeleteBookResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) RestoreBook(ctx context.Context, request *connect.Request[booksv1.RestoreBookRequest]) (*connect.Response[booksv1.RestoreBookResponse], error) {
	book, err := s.service.Restore(ctx, request.Msg.GetBookUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.RestoreBookResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) PurgeBook(ctx context.Context, request *connect.Request[booksv1.PurgeBookRequest]) (*connect.Response[booksv1.PurgeBookResponse], error) {
	if err := s.service.Purge(ctx, request.Msg.GetBookUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.PurgeBookResponse{}), nil
}

func (s *bookServer) EmptyBookTrash(ctx context.Context, _ *connect.Request[booksv1.EmptyBookTrashRequest]) (*connect.Response[booksv1.EmptyBookTrashResponse], error) {
	if err := s.service.EmptyTrash(ctx); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.EmptyBookTrashResponse{}), nil
}

func (s *bookServer) RetryBookProcessing(ctx context.Context, request *connect.Request[booksv1.RetryBookProcessingRequest]) (*connect.Response[booksv1.RetryBookProcessingResponse], error) {
	book, err := s.service.RetryProcessing(ctx, request.Msg.GetBookUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.RetryBookProcessingResponse{Book: bookProto(book)}), nil
}

func (s *bookServer) GetReadingProgress(ctx context.Context, request *connect.Request[booksv1.GetReadingProgressRequest]) (*connect.Response[booksv1.GetReadingProgressResponse], error) {
	progress, err := s.service.GetProgress(ctx, request.Msg.GetBookUid())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.GetReadingProgressResponse{ReadingProgress: readingProgressProto(progress)}), nil
}

func (s *bookServer) UpdateReadingProgress(ctx context.Context, request *connect.Request[booksv1.UpdateReadingProgressRequest]) (*connect.Response[booksv1.UpdateReadingProgressResponse], error) {
	if request.Msg.GetReadingProgress() == nil {
		return nil, connectError(domain.ErrInvalidArgument)
	}
	progress, err := s.service.UpdateProgress(ctx, readingProgressDomain(request.Msg.GetReadingProgress()))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.UpdateReadingProgressResponse{ReadingProgress: readingProgressProto(progress)}), nil
}

func (s *bookServer) ListShelves(ctx context.Context, _ *connect.Request[booksv1.ListShelvesRequest]) (*connect.Response[booksv1.ListShelvesResponse], error) {
	shelves, err := s.service.ListShelves(ctx)
	if err != nil {
		return nil, connectError(err)
	}
	result := make([]*booksv1.Shelf, 0, len(shelves))
	for _, shelf := range shelves {
		result = append(result, shelfProto(shelf))
	}
	return connect.NewResponse(&booksv1.ListShelvesResponse{Shelves: result}), nil
}

func (s *bookServer) CreateShelf(ctx context.Context, request *connect.Request[booksv1.CreateShelfRequest]) (*connect.Response[booksv1.CreateShelfResponse], error) {
	shelf, err := s.service.CreateShelf(ctx, request.Msg.GetDisplayName())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.CreateShelfResponse{Shelf: shelfProto(shelf)}), nil
}

func (s *bookServer) DeleteShelf(ctx context.Context, request *connect.Request[booksv1.DeleteShelfRequest]) (*connect.Response[booksv1.DeleteShelfResponse], error) {
	if err := s.service.DeleteShelf(ctx, request.Msg.GetShelfUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.DeleteShelfResponse{}), nil
}

func (s *bookServer) AddBookToShelf(ctx context.Context, request *connect.Request[booksv1.AddBookToShelfRequest]) (*connect.Response[booksv1.AddBookToShelfResponse], error) {
	if err := s.service.AddToShelf(ctx, request.Msg.GetShelfUid(), request.Msg.GetBookUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.AddBookToShelfResponse{}), nil
}

func (s *bookServer) RemoveBookFromShelf(ctx context.Context, request *connect.Request[booksv1.RemoveBookFromShelfRequest]) (*connect.Response[booksv1.RemoveBookFromShelfResponse], error) {
	if err := s.service.RemoveFromShelf(ctx, request.Msg.GetShelfUid(), request.Msg.GetBookUid()); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&booksv1.RemoveBookFromShelfResponse{}), nil
}
