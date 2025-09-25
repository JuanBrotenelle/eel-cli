import eel
import os


def is_truthy(value: str | None) -> bool:
    if value is None:
        return False
    return value.strip().lower() in {"1", "true", "yes", "on"}


def get_web_root() -> str:
    base_dir = os.path.dirname(__file__)
    web_dir = os.path.join(base_dir, ".distweb")
    if os.path.isdir(web_dir):
        return web_dir


def main() -> None:
    vite_url = os.getenv("VITE_DEV_SERVER_URL")
    if vite_url:
        base_dir = os.path.dirname(__file__)
        dev_dir = os.path.join(base_dir, ".dev_eel")
        os.makedirs(dev_dir, exist_ok=True)
        redirect_html = os.path.join(dev_dir, "index.html")
        with open(redirect_html, "w", encoding="utf-8") as f:
            f.write(
                """<!doctype html><html><head><meta charset=\"utf-8\">
<meta http-equiv=\"refresh\" content=\"0; url=VITE_URL\">
<script>location.replace('VITE_URL');</script>
</head><body>Redirecting to Vite dev server…</body></html>""".replace("VITE_URL", vite_url)
            )
        eel.init(dev_dir)
        eel.start("index.html", size=(1000, 700), port=0)
        return

    web_dir = get_web_root()
    eel.init(web_dir)
    eel.start("index.html", size=(1000, 700), port=0)

if __name__ == '__main__':
    main()