from bottle import route, template, run, request

class DB:
    def __init__(self):
        pass

    def read_messages(self):
        return [
            {"author": "rob", "title": "First post"},
            {"author": "rob", "title": "Second post"},
            {"author": "rob", "title": "Trying out this new bulletin board"}
        ]


class Writer:
    def __init__(self):
        self.lines = []

    def write(self, s):
        self.lines.append(s)

    def text(self):
        return "\n".join(self.lines)


def messages_section(db, **kwargs):
    user = kwargs.get("user", "Guest")
    group = kwargs.get("group") or "(No Group)"

    w = Writer()
    w.write("<pre>")
    w.write(f"<h1>Group: {group}</h1>")
    w.write(f"<h2>Hello {user}</h2>")
    w.write("Latest messages:")
    w.write("")
    w.write("index  author      title    replies  date")
    w.write("-------------------------------------------")

    for msg in db.read_messages():
        w.write(f"{msg['author']}  {msg['title']}")

    w.write("</pre>")

    return w.text()


@route("/")
@route("/<user>")
def cb(user="guest"):
    db = DB()
    #group = request.params.get("group", "No Group")
    group = request.params.group
    return messages_section(db, user=user, group=group)

run(host="localhost", port=8000, debug=True)

