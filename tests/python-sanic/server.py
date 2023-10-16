from sanic import Sanic
from sanic.response import text

app = Sanic("proxied_example")
app.config.FORWARDED_SECRET = "YOUR SECRET"

@app.get("/zeabur")
def index(request):
    return text("zeabur")

if __name__ == "__main__":
    app.run(host="127.0.0.1", port=8000, workers=8, access_log=False)
