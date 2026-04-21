# -*- coding: utf-8 -*-
"""
URL Scraper for inCompass URL Lookup
Automatically retrieves _token and g-recaptcha-response fields
"""

import time
import pandas as pd
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
import warnings

warnings.filterwarnings('ignore')


class URLScraper:
    def __init__(self, headless=True):
        self.headless = headless
        self.driver = None
        self.target_url = "https://incompass.netstar-inc.com/urlsearch"
        self.recaptcha_site_key = "6LfvsHEjAAAAAKEzCUYC281tTl6n9ZsvETRdIYoQ"
    
    def init_driver(self):
        chrome_options = Options()
        if self.headless:
            chrome_options.add_argument("--headless=new")
        chrome_options.add_argument("--no-sandbox")
        chrome_options.add_argument("--disable-dev-shm-usage")
        chrome_options.add_argument("--disable-gpu")
        chrome_options.add_argument("--window-size=1920,1080")
        chrome_options.add_argument("--disable-blink-features=AutomationControlled")
        chrome_options.add_experimental_option("excludeSwitches", ["enable-automation"])
        chrome_options.add_experimental_option('useAutomationExtension', False)
        
        self.driver = webdriver.Chrome(options=chrome_options)
        self.driver.execute_cdp_cmd("Page.addScriptToEvaluateOnNewDocument", {
            "source": """
                Object.defineProperty(navigator, 'webdriver', {
                    get: () => undefined
                })
            """
        })
    
    def get_tokens(self, query_url):
        """
        Get _token and g-recaptcha-response for a given URL
        """
        try:
            if self.driver is None:
                self.init_driver()
            
            self.driver.get(self.target_url)
            time.sleep(2)
            
            wait = WebDriverWait(self.driver, 10)
            
            _token = self.driver.find_element(By.NAME, "_token").get_attribute("value")
            
            self.driver.execute_script(f"""
                var recaptchaToken = document.getElementById('recaptchaToken');
                if (!recaptchaToken) {{
                    var input = document.createElement('input');
                    input.type = 'hidden';
                    input.name = 'g-recaptcha-response';
                    input.id = 'recaptchaToken';
                    document.forms[0].appendChild(input);
                }}
            """)
            
            token = self.driver.execute_async_script(f"""
                var callback = arguments[arguments.length - 1];
                grecaptcha.ready(function() {{
                    grecaptcha.execute('{self.recaptcha_site_key}', {{action: 'submit'}}).then(function(token) {{
                        callback(token);
                    }});
                }});
            """)
            
            self.driver.execute_script(f"""
                document.getElementById('recaptchaToken').value = arguments[0];
            """, token)
            
            g_recaptcha_response = token
            
            return {
                'url': query_url,
                '_token': _token,
                'g-recaptcha-response': g_recaptcha_response,
                'status': 'success'
            }
            
        except Exception as e:
            return {
                'url': query_url,
                '_token': '',
                'g-recaptcha-response': '',
                'status': f'error: {str(e)}'
            }
    
    def process_excel(self, input_file, output_file):
        """
        Process URLs from Excel file and save results
        """
        df = pd.read_excel(input_file)
        
        if '要查询的网址' not in df.columns:
            first_column = df.columns[0]
            urls = df[first_column].tolist()
        else:
            urls = df['要查询的网址'].tolist()
        
        results = []
        
        try:
            self.init_driver()
            
            for i, url in enumerate(urls):
                print(f"Processing {i+1}/{len(urls)}: {url}")
                result = self.get_tokens(url)
                results.append(result)
                time.sleep(2)
                
        finally:
            if self.driver:
                self.driver.quit()
        
        output_df = pd.DataFrame(results)
        output_df = output_df[['url', '_token', 'g-recaptcha-response', 'status']]
        output_df.columns = ['要查询的网址', '_token字段', 'g-recaptcha-response字段', '状态']
        output_df.to_excel(output_file, index=False)
        print(f"\nResults saved to: {output_file}")
        
        return output_df


def main():
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python url_scraper.py <input_excel_file> [output_excel_file]")
        print("Example: python url_scraper.py urls.xlsx results.xlsx")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else "output_results.xlsx"
    
    scraper = URLScraper(headless=True)
    scraper.process_excel(input_file, output_file)


if __name__ == "__main__":
    main()