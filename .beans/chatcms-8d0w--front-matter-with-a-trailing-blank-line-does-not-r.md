---
# chatcms-8d0w
title: Front matter with a trailing blank line does not round trip
status: todo
type: task
priority: normal
created_at: 2026-11-03T20:40:45Z
updated_at: 2026-11-03T20:40:45Z
parent: chatcms-0fml
---

Found by the first recording from the reference instance (chatcms-37w8).

A document whose front matter ends with a blank line, and whose body ends
with blank lines, does not survive a no-op round trip:

    slider: 440
    <blank>
    ---

Reading `test/fixtures/reference/string-in-front-matter/site/dataFields.md`
and writing it back with no edit gives a file one byte longer: the blank line
at the end of the front matter is dropped, and two newlines are added to the
end of the body.

`content/serialisation` says "A document that is read and written back is
unchanged", so this is a requirement that does not hold for a real file from
a real template. It is invisible today because no fixture has that shape.

It is not urgent: the old Quiqr does the same two things when it saves, which
is why the recorded case passes. A customer's repository therefore gets the
same bytes either way. What is wrong is the claim, and a second application
agreeing with a wobble does not make it not a wobble.

What to decide: whether the writer should keep both, or whether the
requirement should say what actually happens and why it is safe. Either way
there should be a fixture with this shape, so the answer stays true.
