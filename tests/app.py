from flask import Flask
import time

app = Flask(__name__)

@app.route('/<timeout>')
def handler(timeout):
    timeout = int(timeout)
    time.sleep(timeout)
    return f'I slept {timeout} seconds'

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)

