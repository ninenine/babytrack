import { useLiveQuery } from 'dexie-react-hooks'
import { db } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'

export function NotesList() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const notes = useLiveQuery(
    () =>
      currentChild
        ? db.notes
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('id')
        : [],
    [currentChild?.id]
  )

  const pinnedNotes = notes?.filter((n) => n.pinned) || []
  const unpinnedNotes = notes?.filter((n) => !n.pinned) || []

  return (
    <div style={{ padding: '1rem' }}>
      <header style={{ marginBottom: '1.5rem' }}>
        <h1>Notes</h1>
        {currentChild && <p>Notes for {currentChild.name}</p>}
      </header>

      <button
        style={{
          backgroundColor: 'var(--primary)',
          color: 'white',
          border: 'none',
          padding: '0.75rem 1.5rem',
          borderRadius: '0.5rem',
          cursor: 'pointer',
          marginBottom: '1rem',
        }}
      >
        + New Note
      </button>

      <div>
        {pinnedNotes.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h3 style={{ marginBottom: '0.5rem' }}>Pinned</h3>
            {pinnedNotes.map((note) => (
              <NoteCard key={note.id} note={note} />
            ))}
          </div>
        )}

        {unpinnedNotes.length === 0 && pinnedNotes.length === 0 ? (
          <p>No notes yet.</p>
        ) : (
          unpinnedNotes.map((note) => <NoteCard key={note.id} note={note} />)
        )}
      </div>
    </div>
  )
}

function NoteCard({
  note,
}: {
  note: { id: string; title?: string; content: string; tags?: string[]; pinned: boolean }
}) {
  return (
    <div
      style={{
        border: '1px solid var(--border)',
        borderRadius: '0.5rem',
        padding: '1rem',
        marginBottom: '0.5rem',
      }}
    >
      {note.title && (
        <div style={{ fontWeight: 'bold', marginBottom: '0.5rem' }}>
          {note.pinned && '📌 '}
          {note.title}
        </div>
      )}
      <div style={{ whiteSpace: 'pre-wrap' }}>{note.content}</div>
      {note.tags && note.tags.length > 0 && (
        <div style={{ marginTop: '0.5rem' }}>
          {note.tags.map((tag) => (
            <span
              key={tag}
              style={{
                backgroundColor: 'var(--surface)',
                padding: '0.125rem 0.5rem',
                borderRadius: '0.25rem',
                marginRight: '0.25rem',
                fontSize: '0.75rem',
              }}
            >
              #{tag}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
