This is my submision to the third golang challenge. I started when it first came out so I have done many of the requirements that are no longer required.

Things I learned:

1. Oauth. I implemented oauth authentication to imgur.com. I intended to have a user select from their images and upload back to their account, but once the requirements were lessened, I didn't get back to fully completing that integration. I do however, rely on imgur's thumbnailing capabilities to give nice 90x90 thumbnails for the web-app's built in collections.

2. Security. I was concerned about security of oauth tokens, but also didn't want to require any backend for my web app. I implemented a secure cookie scheme that I rather liked.

3. Web. I developed a novel system for managing and rendering templates that I am rather proud of. The web flow is still probably the most convenient way to generate mosaics, unless youy want to supply your own thumbnails. 

4. Concurrency. One of my goals for the web app was to be responsive to the user even though mosaic generating may take a "long time" (at least longer than is acceptable to wait for a web request.) I engineered a smart queue and status notification system using long polling and some cool channel tricks.

5. Images. Mosaics were a new challenge to me. I impemented a naive evaluation function that simply averages the entire thumbnail. I also implemented a mode that splits target areas into a nxn grid and evaluates each portion seperately. That worked, but as a colorblind person, I really couldn't tell much difference. It benchmarked a bit slower, so I prever the basic Averageing Evaluator.