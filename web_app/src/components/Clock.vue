<template>
  <div class="clock-container">
    <div class="button-row">
      <button
        class="uk-button uk-button-default"
        v-on:click="clockStartAndRefetch('10m', true)"
      >
        10 (work)
      </button>
      <button
        class="uk-button uk-button-primary"
        v-on:click="clockStartAndRefetch('10m', false)"
      >
        10 (spare)
      </button>
    </div>
    <div class="button-row">
      <button
        class="uk-button uk-button-default"
        v-on:click="clockStartAndRefetch('25m', true)"
      >
        25 (work)
      </button>
      <button
        class="uk-button uk-button-primary"
        v-on:click="clockStartAndRefetch('25m', false)"
      >
        25 (spare)
      </button>
    </div>
  </div>

  <hr class="uk-divider-icon" />

  <div class="ongoing-clock-container">
    <div v-for="clock in clocks" v-bind:key="clock" class="ongoing-clock-row">
      <span class="">{{ clock.Minute }} : {{ clock.Second }} </span>
      <span v-bind:class="'clock-label-text clock-label-text-' + clock.Tag">
        {{ clock.Tag }}
      </span>
      <button
        class="uk-button uk-button-primary"
        v-on:click="clockStop(clock.Id)"
      >
        stop
      </button>
    </div>
  </div>

  <div class="uk-child-width-1-3@m uk-child-width-1-2@s records-container">
    <div v-for="item in items" v-bind:key="item" class="record-card-container">
      <div class="uk-card uk-card-default uk-card-body">
        <h3 class="uk-card-title">
          <div class="uk-card-badge uk-label">{{ item.Count }}</div>
          {{ item.Title }}
        </h3>
        <table class="uk-table uk-table-small uk-text-nowrap">
          <tbody>
            <tr v-for="readable in item.Readables" v-bind:key="readable">
              <td class="uk-width-1-4">{{ readable.Start }}</td>
              <td class="uk-width-1-4">{{ readable.Duration }}</td>
              <td v-bind:class="'clock-label-text-' + readable.Tag">
                {{ readable.Tag }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import { ref, onMounted } from "vue";
const apiUrl = "http://localhost:8000";

function stopWatchReadingSet(clock, now) {
  let start = new Date(clock.StartTime);
  let remain = (start - now + clock.Duration / 1000000) / 1000;
  clock.Minute = Math.max(Math.floor((remain % 3600) / 60), 0);
  clock.Second = Math.max(Math.floor(remain - 60 * clock.Minute), 0);
}

function fetchClocks() {
  return axios
    .get(apiUrl + "/api/clocks")
    .then((resp) => (resp.data ? resp.data : []))
    .then((clocks) => {
      let now = new Date();
      clocks.forEach((clock) => {
        stopWatchReadingSet(clock, now);
      });
      return clocks;
    });
}

export default {
  setup() {
    const items = ref([]);
    const clocks = ref([]);

    const getRecords = async () => {
      try {
        items.value = await axios
          .get(apiUrl + "/api/records")
          .then((resp) => resp.data.Items);
      } catch (e) {
        console.log(e);
      }
    };
    const getClocks = async () => {
      try {
        clocks.value = await fetchClocks();
      } catch (e) {
        console.log(e);
      }
    };
    const countDown = () => {
      setInterval(() => {
        if (!clocks.value) {
          return;
        }
        let now = new Date();
        clocks.value = clocks.value.filter((clock) => {
          stopWatchReadingSet(clock, now);
          if (clock.Minute == 0 && clock.Second == 0) {
            getRecords();
            return false;
          }
          return true;
        });
      }, 500);
    };

    onMounted(() => {
      getRecords();
      getClocks();
      countDown();
    });

    return {
      items,
      clocks,
    };
  },
  computed: {},
  methods: {
    async clockStart(time, forWork) {
      try {
        await axios.post(apiUrl + "/tomato", {
          ctlStr: (forWork ? "w" : "s") + " " + time,
        });
      } catch (e) {
        console.log(e);
      }
    },
    async sleep(time) {
      await new Promise((resolve) => setTimeout(resolve, time));
    },
    async runningClockGet() {
      try {
        return await fetchClocks();
      } catch (e) {
        console.log(e);
      }
    },
    async clockStartAndRefetch(time, forWork) {
      this.clocks = await this.clockStart(time, forWork).then(() =>
        this.runningClockGet()
      );
    },
    async clockStop(id) {
      try {
        this.clocks = await axios
          .post(apiUrl + "/api/clockStop", {
            id: id,
          })
          .then(() => this.runningClockGet());
      } catch (e) {
        console.log(e);
      }
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.clock-container {
  padding: 40px 20px 0px 20px;
  display: flex;
  flex-direction: column;
}

.button-row {
  display: flex;
  justify-content: center;
  padding: 10px;
  gap: 10px;
}

.ongoing-clock-container {
  padding: 0px 25px 25px;
}

.uk-countdown-number {
  font-size: 1rem;
}

.ongoing-clock-row {
  font-size: 1rem;
  display: flex;
  padding: 10px;
  gap: 25px;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #e5e5e5;
}

.record-card-container {
  padding: 25px;
}

.records-container {
  display: flex;
  align-items: center;
  flex-direction: column;
}

.clock-label-text {
  text-align: center;
  width: 40px;
}
.clock-label-text-work {
  color: #1e87f0;
}
.clock-label-text-spare {
  color: #32d296;
}
</style>
