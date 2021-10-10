# Backend Translations

PhotoPrism uses [gettext](https://en.wikipedia.org/wiki/Gettext) for localizing frontend and backend.
It's one of the most widely adopted standards for translating user interfaces.
 
Human-readable messages like `File not found` are used as ids for finding matching translations, 
and used as defaults whenever there is no translation available.

Messages may optionally contain placeholders, like `Found %d files`, for numbers and 
other variables.

We strongly recommend [Poedit](https://poedit.net/download) for creating and updating translations.
Download is free for Mac, Windows, and Linux.
It's source code can be obtained on [GitHub](https://github.com/vslavik/poedit).

Only asynchronous notifications and certain API responses need translation to provide a 
consistent user experience.
Technical log messages should be in English to avoid ambiguities and (even slightly) wrong translations. 

`default.po` files in subdirectories contain localized messages for each 
[language](https://www.gnu.org/software/gettext/manual/html_node/Usual-Language-Codes.html) identified 
by their [locale](https://www.gnu.org/software/gettext/manual/html_node/Locale-Names.html),
for example `de/default.po` for German and `pt_BR/default.po` for Brazilian Portuguese. 
You can open, edit, and save them with Poedit. Please also add and commit binary `*.mo` files, 
which will be automatically created by Poedit.

To add a new translation, open `messages.pot`, click on "Create New Translation" at the bottom and select
the language. Now you can start translating.
When done, create a new directory (using the locale as name) and save your translation there as `default.po`.

The POT file `/assets/locales/messages.pot` will be automatically updated when 
running `go generate` in `/internal/i18n` or `make generate` in the main project directory.
Note that this will only work when you have gettext installed on your system.
We recommend using our latest development image as described in the Developer Guide.

Apply changes to existing translations by clicking on "Catalogue" > "Update from POT File..." 
in the Poedit app menu.

