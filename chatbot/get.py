import requests
import re
import json

with open('secrets.json', 'r') as secretsfile:
    secrets = json.load(secretsfile)
    slack_key = secrets['slack-key']

with open('test.txt', 'a') as myfile:
    for page in range(1,41):
        try:
            r = requests.get('https://slack.com/api/search.messages?token={}&query=from%3A%40jackson.merkl%20-%22shared%20a%20file%3A%22%20-%22uploaded%20a%20file%3A%22&count=40&page={}&pretty=1'.format(slack_key, page))
            for message in r.json()['messages']['matches']:
                if not 'http' in message['text'] and message['channel']['name'] != 'team-leadership' and message['channel']['name'] != 'electrical-leadership':
                    new_message = re.sub(r'<!channel> ', '', message['text'])
                    new_message = re.sub(r'\<.*?\>:', '', new_message)
                    new_message = re.sub(r'\<.*?\>', '', new_message)
                    print(new_message)
                    myfile.write(new_message + "\n")
        except KeyError:
            print(r.json())
            print(page)
            raise