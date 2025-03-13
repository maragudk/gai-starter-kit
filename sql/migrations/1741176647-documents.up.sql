create table documents (
  id text primary key default ('d_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  content text not null
) strict;

create trigger documents_updated_timestamp after update on documents begin
  update documents set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create table chunks (
  id text primary key default ('c_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  documentID text not null references documents (id) on delete cascade,
  "index" int not null,
  content text not null
) strict;

create trigger chunks_updated_timestamp after update on chunks begin
  update chunks set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create index chunks_documentID_index_index on chunks (documentID, "index");

-- create a virtual table for full-text search using the porter tokenizer
create virtual table chunks_fts using fts5(
  content, tokenize = porter, content = 'chunks'
);

create trigger chunks_after_insert after insert on chunks begin
  insert into chunks_fts (rowid, content) values (new.rowid, new.content);
end;

create trigger chunks_fts_after_update after update on chunks begin
  insert into chunks_fts (chunks_fts, rowid, content) values('delete', old.rowid, old.content);
  insert into chunks_fts (rowid, content) values (new.rowid, new.content);
end;

create trigger chunks_fts_after_delete after delete on chunks begin
  insert into chunks_fts (chunks_fts, rowid, content) values('delete', old.rowid, old.content);
end;

create virtual table chunk_embeddings using vec0(
  chunkID text primary key references chunks (id) on delete cascade,
  embedding float[1024]
);
