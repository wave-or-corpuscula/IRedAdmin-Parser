from typing import Dict, Optional

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager


class WebAuthService:

    def __init__(self, driver_path: Optional[str] = None) -> None:
        self.driver_path = driver_path

    def open_user_page(self, url: str, cookie: Dict[str, str]) -> None:
        options = Options()
        options.add_experimental_option("detach", True)
        options.add_argument("--no-sandbox")
        options.add_argument("--disable-dev-shm-usage")

        if self.driver_path:
            service = Service(self.driver_path)
            driver = webdriver.Chrome(service=service, options=options)
        else:
            driver = webdriver.Chrome(
                service=Service(ChromeDriverManager().install()),
                options=options
            )

        try:
            driver.get(url)

            driver.delete_all_cookies()
            driver.execute_cdp_cmd("Network.setCookie", cookie)

            driver.refresh()
        except Exception:
            driver.quit()
            raise

