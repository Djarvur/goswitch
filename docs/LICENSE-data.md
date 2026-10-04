# Data licenses — baked detector dictionaries

The autocorrect detector's dictionary data (D-52, plan 06-02) is derived from
the hunspell dictionaries installed on the dev machine and baked as committed
golden Go sources by `layouts/dictgen`:

- `layouts/dict_ru.go` — `DictRU` (hunspell ru_RU)
- `layouts/dict_en.go` — `DictEN` (hunspell en_US)
- `layouts/trigrams.go` — `TriRU`/`TriEN` (trained on the same two word lists)

Baked: 2026-10-01, from hunspell-ru 1:24.2.1-1 and hunspell-en-us
1:2020.12.07-2 (Ubuntu 24.04). The generator (`go generate ./layouts`) is the
only path the data may travel; the verbatim upstream notices are preserved
below per the licensing verdict table of 06-RESEARCH (Q3).

## Russian — ru_RU

Source file: `/usr/share/hunspell/ru_RU.dic` (Debian package `hunspell-ru`,
upstream libreoffice-dictionaries). 146 269 entries; after the generator's
affix-flag stripping, minimum-length filter and deduplication: 146 261 unique
words; with the mandated ё→е folding (Pitfall 3 / R3): 138 914.

Verbatim `Files:` block from `/usr/share/doc/hunspell-ru/copyright`:

    Files: dictionaries/ru_RU/*
    Copyright: 1997-2008 Alexander I. Lebedev
    License: custom-bsd-4-clauses
    All rights reserved.
    .
    Redistribution and use in source and binary forms, with or without
    modification, are permitted provided that the following conditions
    are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Modified versions must be clearly marked as such.
    * The name of Alexander I. Lebedev may not be used to endorse or promote
      products derived from this software without specific prior written
      permission.
    .
    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
    AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
    IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
    ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
    LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
    SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
    INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
    CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
    ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
    POSSIBILITY OF SUCH DAMAGE.

## English — en_US

Source file: `/usr/share/hunspell/en_US.dic` (Debian package `hunspell-en-us`,
upstream SCOWL). 79 013 entries; after the generator's processing: 78 951
unique words. Verbatim excerpt from `/usr/share/doc/hunspell-en-us/copyright`:

    SCOWL (Spell Checker Oriented Word Lists) is a collection of
    English word lists maintained by Kevin Atkinson
    <kevina@users.sourceforge.net>

    Copyright: (extracted from the SCOWL README file):

    The collective work is Copyright 2000-2011 by Kevin Atkinson as well
    as any of the copyrights mentioned below:

      Copyright 2000-2011 by Kevin Atkinson

      Permission to use, copy, modify, distribute and sell these word
      lists, the associated scripts, the output created from the scripts,
      and its documentation for any purpose is hereby granted without fee,
      provided that the above copyright notice appears in all copies and
      that both that copyright notice and this permission notice appear in
      supporting documentation. Kevin Atkinson makes no representations
      about the suitability of this array for any purpose. It is provided
      "as is" without express or implied warranty.

The 10 level of the list includes material from the Moby (TM) Words II
[MWords] package, explicitly placed in the public domain by Grady Ward, and
from Brian Kelk's "UK English Wordlist with Frequency Classification", also
in the public domain; Alan Beale (12Dicts package, ENABLE) contributed
further public-domain material. The full verbatim text of these notices is
preserved in `/usr/share/doc/hunspell-en-us/copyright`.

## Prohibited sources

Frequency lists from OpenCorpora are NOT used (license never verified —
06-RESEARCH A1); the baked dictionaries are the FULL hunspell word lists. No
code was copied from the GPL donors (xneur, Easy Switcher — read-only
algorithmic references, ADR-007); `layouts/dictgen` is written from scratch.
