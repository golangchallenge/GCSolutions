License BSD.

Known issues:
 * Some unit tests are not passing, due to blocking io (io.ReadFull).
   Implemented Read methods in boxReader, Conn are greedy.
   Quote from io.Reader:
   If some data is available but not len(p) bytes, Read conventionally returns
   what is available instead of waiting for more.
