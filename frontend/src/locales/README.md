# Frontend Translations

PhotoPrism uses [gettext](https://en.wikipedia.org/wiki/Gettext) for localizing frontend and backend.
It's one of the most widely adopted standards for translating user interfaces.
 
Human-readable messages like `File not found` are used as ids for finding matching translations, 
and used as defaults whenever there is no translation available.

Messages may optionally contain placeholders, like `%{n} files found`, for numbers and 
other variables.

We strongly recommend [Poedit](https://poedit.net/download) for creating and updating translations.
Download is free for Mac, Windows, and Linux.
It's source code can be obtained on [GitHub](https://github.com/vslavik/poedit).

`*.po` files contain localized messages for each 
[language](https://www.gnu.org/software/gettext/manual/html_node/Usual-Language-Codes.html) identified 
by their [locale](https://www.gnu.org/software/gettext/manual/html_node/Locale-Names.html),
for example `de.po` for German and `pt_BR.po` for Brazilian Portuguese.
You can open, edit, and save them with Poedit to update existing translations. 

To add a new translation, open `translations.pot`, click on "Create New Translation" at the bottom and select
the language. Now you can start translating. 
When done, save your translation as `*.po` file using the [locale](https://www.gnu.org/software/gettext/manual/html_node/Locale-Names.html) as name.
In addition, the new language needs to be added to the `Languages` function in `/frontend/src/options/options.js`.

A binary `*.mo` (machine object) file will be automatically saved along with every `*.po` file. 
You won't be able to open those in a text editor, but please include them in git commits or when sending
translations via email. The compiled `translations.json` file is not required for pull requests 
and often causes merge conflicts.

If you have a working development environment in place:

Running `npm run gettext-compile` in the `frontend` directory compiles existing translations into 
a single `translations.json` file.

Now start a frontend build using `npm run build` or keep 

```
npm run watch
```

running in the background to automatically recompile JS and CSS whenever there
are changes. Lastly, make sure `photoprism` is running and open the Web UI in a supported browser. Changing 
the language in Settings automatically triggers a reload.

To extract new or changed text needing translation from `*.js` and `*.vue` source code, run 

```
npm run gettext-extract
```

in the `frontend` directory. This updates the POT file `translations.pot`.

Apply changes to existing translations by clicking on "Catalogue" > "Update from POT File..." 
in the Poedit app menu.

