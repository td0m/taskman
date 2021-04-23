# UI Concepts

## Outline view
This view shows all tasks within a given timeline, meaning they satisfy any of the following:
 - not completed
 - completed today, yesterday, or two days ago

<pre><code><div style="display: flex; justify-content: space-between"><div><b>Outline</b> | Today</div><div>20 Apr</div><div>         10%</div></div>
<b>Coming up</b>
  4 days    PLC Coursework
  5 days    SEG 4 Deliverable
  2nd May   SEG Meeting at 4pm
  ... <!-- this gets trimmed if tasks won't fit on the screen -->

<b>SEG Coursework</b> <span style="opacity: 0.7">(3)</span>
 ∙ <strike style="opacity: 0.7">Group meeting at 4pm</strike>
 ∙ <strike style="opacity: 0.7">Merge Histogram PR</strike>
 ∙ Integrate Filters into Histogram
    ∙ <strike style="opacity: 0.7">Send PR</strike>
    ∙ Assign to Anan for review
 ∙ <strike style="opacity: 0.7">Ask if document is ready</strike>
 ∙ Submit document (when ready)
</code></pre>

 ^ Note that this does not mean that SEG Coursework only has 6 tasks. Finished tasks automatically get hidden from outline if they were completed 3 or more days ago. You'll notice that the first subtask doesn't exist in 

## Today view
 - due today or overdue
<pre><code><div style="display: flex; justify-content: space-between"><div>Outline | <b>Today</b></div><div>20 Apr</div><div>          50%</div></div>
<b>SEG Coursework</b> <span style="opacity: 0.7">(3/6)</span>
 ∙ <strike style="opacity: 0.7">Merge Histogram PR</strike>
 ∙ Integrate Filters into Histogram
    ∙ <strike style="opacity: 0.7">Send PR</strike>
    ∙ Assign to Anan for review
 ∙ <strike style="opacity: 0.7">Ask if document is ready</strike>
 ∙ Submit document (when ready)
</code></pre>

## Time travel

You can also time travel:
<pre><code><div style="display: flex; justify-content: center"><i style="color: red">11th Mar</i></div>

<div style="display: flex;justify-content: center; flex-direction: column; align-items: center;"><center style="font-weight: 600"><  January 2021  ></center><div>02 03 04 05 06 07 08
09 10 11 12 13 14 15
16 17 18 19 20 21 22
23 24 25 26 27 <b>28</b> 29
30 31</div>
<div>
<b>at</b> 6:30am</div>
</div>
</code></pre>

This allows you to view tasks the way they were any time before (or will look in the future). This also allows you to create tasks that will only be visible in the future, or see what the schedule looked like before creating a bunch of tasks on one day. It can also be used for clock ins.