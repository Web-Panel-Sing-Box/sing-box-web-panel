from app.models import Inbound, User
from app.services.link_builder import build_user_link


def test_build_vless_reality_link() -> None:
    inbound = Inbound(
        id=1,
        protocol="vless",
        tag="reality",
        listen="::",
        port=443,
        tls_enabled=True,
        reality_enabled=True,
        server_name="example.com",
        reality_public_key="pub",
        reality_short_id="abcd",
    )
    user = User(id=1, inbound_id=1, username="alice", uuid="00000000-0000-0000-0000-000000000000")

    link = build_user_link(user, inbound, public_host="vpn.example.com")

    assert link.startswith("vless://00000000-0000-0000-0000-000000000000@vpn.example.com:443?")
    assert "security=reality" in link
    assert "pbk=pub" in link
    assert link.endswith("#alice")
