package sql

import (
	"app/model"
	"context"
	"fmt"
	"strings"

	"maragu.dev/errors"
)

// Search chunks that match the query and embedding. Matches using FTS first, then vector similarity search.
// See https://alexgarcia.xyz/blog/2024/sqlite-vec-hybrid-search/ for the search query.
func (d *Database) Search(ctx context.Context, q string, embedding []byte) ([]model.Chunk, error) {
	// Do exact matches only in FTS for now
	q = fmt.Sprintf(`"%v"`, strings.Trim(q, `"`))

	query := `
		with

		fts_matches as (
			select
				id,
				row_number() over (order by bm25(chunks_fts)) as rank_number,
				1 as source
			from chunks
				join chunks_fts on (chunks.rowid = chunks_fts.rowid)
			where chunks_fts.content match ?
			order by bm25(chunks_fts)
		),

		vector_matches as (
			select
				id,
				row_number() over (order by distance) as rank_number,
				2 as source
			from chunks
				join chunk_embeddings on (chunks.id = chunk_embeddings.chunkID)
			where
				k = 100 and
				distance < 0.75 and
				embedding match ?
			order by distance
		),

		combined as (
			select id, rank_number, source from fts_matches
			union all
			select id, rank_number, source from vector_matches
		)

		select distinct chunks.*
		from chunks
			join combined using (id)
		order by combined.source, combined.rank_number`

	var chunks []model.Chunk
	if err := d.H.Select(ctx, &chunks, query, q, embedding); err != nil {
		return chunks, errors.Wrap(err, "error searching chunks")
	}
	return chunks, nil
}
