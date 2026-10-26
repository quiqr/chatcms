# Tasks

## 1. Sending

- [x] 1.1 An SMTP sender behind one interface
- [x] 1.2 Settings, each readable from a file as well as a value
- [x] 1.3 Refuse an address or a header that would smuggle another header
- [x] 1.4 A deadline on the whole exchange, not only the dial

## 2. The messages

- [x] 2.1 An invitation that says who invited them and to what
- [x] 2.2 A password reset that says how long it lasts
- [x] 2.3 Word that a password was changed
- [x] 2.4 Plain text and HTML, with no git word in either
- [x] 2.5 A test that the copy rules hold

## 3. On the queue

- [x] 3.1 Post a message rather than sending it in the request
- [x] 3.2 Retry a failure; drop what can never succeed
- [x] 3.3 Register the handler where the queue and mail meet

## 4. With nowhere to send

- [x] 4.1 Write the link to the log instead
- [x] 4.2 Say so when the server starts
- [x] 4.3 Say so again if posting itself fails

## 5. The screens the links need

- [x] 5.1 Choosing a password from an invitation or a reset
- [x] 5.2 Asking for a reset, saying the same thing either way
- [x] 5.3 A link with no token says what to do
- [x] 5.4 Both at 1440 px and at 390 px

## 6. Mailpit

- [x] 6.1 In the development shell
- [x] 6.2 Started and stopped by scripts/dev.sh, its inbox in the banner
- [x] 6.3 One per acceptance run, so tests never read each other's mail
- [x] 6.4 An acceptance test that reads the invitation and follows the link

## 7. Deploying it

- [x] 7.1 The NixOS module's settings, password from a file
- [x] 7.2 docs/DEPLOYING.md says what a real deployment needs
